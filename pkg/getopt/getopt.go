// Package getopt implements a command-line argument parser.
//
// It tries to cover all common styles of option syntaxes, and provides context
// information when given a partial input. It is mainly useful for writing
// completion engines and wrapper programs.
//
// If you are looking for an option parser for your go program, consider using
// the flag package in the standard library instead.
package getopt

//go:generate stringer -type=Config,Arity,ContextType -output=zstring.go

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/errutil"
)

// Config configures the parsing behavior.
type Config uint

const (
	// Stop parsing options after "--".
	StopAfterDoubleDash Config = 1 << iota
	// Stop parsing options before the first non-option argument.
	StopBeforeFirstNonOption
	// Allow long options to start with "-", and disallow short options.
	// Replicates the behavior of getopt_long_only and the flag package.
	LongOnly

	// Config to replicate the behavior of GNU's getopt_long.
	GNU = StopAfterDoubleDash
	// Config to replicate the behavior of BSD's getopt_long.
	BSD = StopAfterDoubleDash | StopBeforeFirstNonOption
)

// Tests whether a configuration has all specified flags set.
func (c Config) has(bits Config) bool { return c&bits == bits }

// OptionSpec is a command-line option.
type OptionSpec struct {
	// Short option. Set to 0 for long-only.
	Short rune
	// Long option. Set to "" for short-only.
	Long string
	// Whether the option takes an argument, and whether it is required.
	Arity Arity
}

// Arity indicates whether an option takes an argument, and whether it is
// required.
type Arity uint

const (
	// The option takes no argument.
	NoArgument Arity = iota
	// The option requires an argument. The argument can come either directly
	// after a short option (-oarg), after a long option followed by an equal
	// sign (--long=arg), or as a separate argument after the option (-o arg,
	// --long arg).
	RequiredArgument
	// The option takes an optional argument. The argument can come either
	// directly after a short option (-oarg) or after a long option followed by
	// an equal sign (--long=arg).
	OptionalArgument
)

// Option represents a parsed option.
type Option struct {
	Spec     *OptionSpec
	Unknown  bool
	Long     bool
	Argument string
}

// Context describes the context of the last argument.
type Context struct {
	// The nature of the context.
	Type ContextType
	// Current option, with a likely incomplete Argument. Non-nil when Type is
	// OptionArgument.
	Option *Option
	// Current partial long option name or argument. Non-empty when Type is
	// LongOption or Argument.
	Text string
}

// ContextType encodes how the last argument can be completed.
type ContextType uint

const (
	// OptionOrArgument indicates that the last element may be either a new
	// option or a new argument. Returned when it is an empty string.
	OptionOrArgument ContextType = iota
	// AnyOption indicates that the last element must be new option, short or
	// long. Returned when it is "-".
	AnyOption
	// LongOption indicates that the last element is a long option (but not its
	// argument). The partial name of the long option is stored in Context.Text.
	LongOption
	// ChainShortOption indicates that a new short option may be chained.
	// Returned when the last element consists of a chain of options that take
	// no arguments.
	ChainShortOption
	// OptionArgument indicates that the last element list must be an argument
	// to an option. The option in question is stored in Context.Option.
	OptionArgument
	// Argument indicates that the last element is a non-option argument. The
	// partial argument is stored in Context.Text.
	Argument
)

// Parse parses an argument list. It returns the parsed options, the non-option
// arguments, and any error.
func Parse(args []string, specs []*OptionSpec, cfg Config) ([]*Option, []string, error) {
	opts, nonOptArgs, opt, _ := parse(args, specs, cfg)
	var err error
	if opt != nil {
		err = fmt.Errorf("missing argument for %s", optionPart(opt))
	}
	for _, opt := range opts {
		if opt.Unknown {
			err = errutil.Multi(err, fmt.Errorf("unknown option %s", optionPart(opt)))
		}
	}
	return opts, nonOptArgs, err
}

func optionPart(opt *Option) string {
	if opt.Long {
		return "--" + opt.Spec.Long
	}
	return "-" + string(opt.Spec.Short)
}

// Complete parses an argument list for completion. It returns the parsed
// options, the non-option arguments, and the context of the last argument. It
// tolerates unknown options, assuming that they take optional arguments.
func Complete(args []string, specs []*OptionSpec, cfg Config) ([]*Option, []string, Context) {
	opts, nonOptArgs, opt, stopOpt := parse(args[:len(args)-1], specs, cfg)

	arg := args[len(args)-1]
	var ctx Context
	switch {
	case opt != nil:
		opt.Argument = arg
		ctx = Context{Type: OptionArgument, Option: opt}
	case stopOpt:
		ctx = Context{Type: Argument, Text: arg}
	case arg == "":
		ctx = Context{Type: OptionOrArgument}
	case arg == "-":
		ctx = Context{Type: AnyOption}
	case strings.HasPrefix(arg, "--"):
		if !strings.ContainsRune(arg, '=') {
			ctx = Context{Type: LongOption, Text: arg[2:]}
		} else {
			newopt, _ := parseLong(arg[2:], specs)
			ctx = Context{Type: OptionArgument, Option: newopt}
		}
	case strings.HasPrefix(arg, "-"):
		if cfg.has(LongOnly) {
			if !strings.ContainsRune(arg, '=') {
				ctx = Context{Type: LongOption, Text: arg[1:]}
			} else {
				newopt, _ := parseLong(arg[1:], specs)
				ctx = Context{Type: OptionArgument, Option: newopt}
			}
		} else {
			newopts, _ := parseShort(arg[1:], specs)
			if newopts[len(newopts)-1].Spec.Arity == NoArgument {
				opts = append(opts, newopts...)
				ctx = Context{Type: ChainShortOption}
			} else {
				opts = append(opts, newopts[:len(newopts)-1]...)
				ctx = Context{Type: OptionArgument, Option: newopts[len(newopts)-1]}
			}
		}
	default:
		ctx = Context{Type: Argument, Text: arg}
	}
	return opts, nonOptArgs, ctx
}

func parse(args []string, spec []*OptionSpec, cfg Config) ([]*Option, []string, *Option, bool) {
	var (
		opts       []*Option
		nonOptArgs []string
		// Non-nil only when the last argument was an option with required
		// argument, but the argument has not been seen.
		opt *Option
		// Whether option parsing has been stopped. The condition is controlled
		// by the StopAfterDoubleDash and StopBeforeFirstNonOption bits in cfg.
		stopOpt bool
	)
	for _, arg := range args {
		switch {
		case opt != nil:
			opt.Argument = arg
			opts = append(opts, opt)
			opt = nil
		case stopOpt:
			nonOptArgs = append(nonOptArgs, arg)
		case cfg.has(StopAfterDoubleDash) && arg == "--":
			stopOpt = true
		case strings.HasPrefix(arg, "--") && arg != "--":
			newopt, needArg := parseLong(arg[2:], spec)
			if needArg {
				opt = newopt
			} else {
				opts = append(opts, newopt)
			}
		case strings.HasPrefix(arg, "-") && arg != "--" && arg != "-":
			if cfg.has(LongOnly) {
				newopt, needArg := parseLong(arg[1:], spec)
				if needArg {
					opt = newopt
				} else {
					opts = append(opts, newopt)
				}
			} else {
				newopts, needArg := parseShort(arg[1:], spec)
				if needArg {
					opts = append(opts, newopts[:len(newopts)-1]...)
					opt = newopts[len(newopts)-1]
				} else {
					opts = append(opts, newopts...)
				}
			}
		default:
			nonOptArgs = append(nonOptArgs, arg)
			if cfg.has(StopBeforeFirstNonOption) {
				stopOpt = true
			}
		}
	}
	return opts, nonOptArgs, opt, stopOpt
}

// Parses short options, without the leading dash. Returns the parsed options
// and whether an argument is still to be seen.
func parseShort(s string, specs []*OptionSpec) ([]*Option, bool) {
	var opts []*Option
	var needArg bool
	for i, r := range s {
		opt := findShort(r, specs)
		if opt != nil {
			if opt.Arity == NoArgument {
				opts = append(opts, &Option{Spec: opt})
				continue
			} else {
				parsed := &Option{Spec: opt, Argument: s[i+len(string(r)):]}
				opts = append(opts, parsed)
				needArg = parsed.Argument == "" && opt.Arity == RequiredArgument
				break
			}
		}
		// Unknown option, treat as taking an optional argument
		parsed := &Option{
			Spec: &OptionSpec{r, "", OptionalArgument}, Unknown: true,
			Argument: s[i+len(string(r)):]}
		opts = append(opts, parsed)
		break
	}
	return opts, needArg
}

func findShort(r rune, specs []*OptionSpec) *OptionSpec {
	for _, opt := range specs {
		if r == opt.Short {
			return opt
		}
	}
	return nil
}

// Parses a long option, without the leading dashes. Returns the parsed option
// and whether an argument is still to be seen.
func parseLong(s string, specs []*OptionSpec) (*Option, bool) {
	eq := strings.IndexRune(s, '=')
	for _, opt := range specs {
		if s == opt.Long {
			return &Option{Spec: opt, Long: true}, opt.Arity == RequiredArgument
		} else if eq != -1 && s[:eq] == opt.Long {
			return &Option{Spec: opt, Long: true, Argument: s[eq+1:]}, false
		}
	}
	// Unknown option, treat as taking an optional argument
	if eq == -1 {
		return &Option{
			Spec: &OptionSpec{0, s, OptionalArgument}, Unknown: true, Long: true}, false
	}
	return &Option{
		Spec: &OptionSpec{0, s[:eq], OptionalArgument}, Unknown: true,
		Long: true, Argument: s[eq+1:]}, false
}
