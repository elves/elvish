// Package getopt implements a command-line argument parser.
//
// It tries to cover all common styles of option syntaxes, and provides context
// information when given a partial input. It is mainly useful for writing
// completion engines and wrapper programs.
//
// If you are looking for an option parser for your go programm, consider using
// the flag package in the standard library instead.
package getopt

//go:generate stringer -type=Config,HasArg,ContextType -output=string.go

import "strings"

// Getopt specifies the syntax of command-line arguments.
type Getopt struct {
	Options []*Option
	Config  Config
}

// Config configurates the parsing behavior.
type Config uint

const (
	// DoubleDashTerminatesOptions indicates that all elements after an argument
	// "--" are treated as arguments.
	DoubleDashTerminatesOptions Config = 1 << iota
	// FirstArgTerminatesOptions indicates that all elements after the first
	// argument are treated as arguments.
	FirstArgTerminatesOptions
	// LongOnly indicates that long options may be started by either one or two
	// dashes, and short options are not allowed. Should replicate the behavior
	// of getopt_long_only and the
	// flag package of the Go standard library.
	LongOnly
	// GNUGetoptLong is a configuration that should replicate the behavior of
	// GNU getopt_long.
	GNUGetoptLong = DoubleDashTerminatesOptions
	// POSIXGetopt is a configuration that should replicate the behavior of
	// POSIX getopt.
	POSIXGetopt = DoubleDashTerminatesOptions | FirstArgTerminatesOptions
)

// HasAll tests whether a configuration has all specified flags set.
func (conf Config) HasAll(flags Config) bool {
	return (conf & flags) == flags
}

// Option is a command-line option.
type Option struct {
	// Short option. Set to 0 for long-only.
	Short rune
	// Long option. Set to "" for short-only.
	Long string
	// Whether the option takes an argument, and whether it is required.
	HasArg HasArg
}

// HasArg indicates whether an option takes an argument, and whether it is
// required.
type HasArg uint

const (
	// NoArgument indicates that an option takes no argument.
	NoArgument HasArg = iota
	// RequiredArgument indicates that an option must take an argument. The
	// argument can come either directly after a short option (-oarg), after a
	// long option followed by an equal sign (--long=arg), or as a subsequent
	// argument after the option (-o arg, --long arg).
	RequiredArgument
	// OptionalArgument indicates that an option takes an optional argument.
	// The argument can come either directly after a short option (-oarg) or
	// after a long option followed by an equal sign (--long=arg).
	OptionalArgument
)

// ParsedOption represents a parsed option.
type ParsedOption struct {
	Option   *Option
	Long     bool
	Argument string
}

// Context indicates what may come after the supplied argument list.
type Context struct {
	// The nature of the context.
	Type ContextType
	// Current option, with a likely incomplete Argument. Non-nil when Type is
	// OptionArgument.
	Option *ParsedOption
	// Current partial long option name or argument. Non-empty when Type is
	// LongOption or Argument.
	Text string
}

// ContextType encodes what may be appended to the last element of the argument
// list.
type ContextType uint

const (
	// NewOptionOrArgument indicates that the last element may be either a new
	// option or a new argument. Returned when it is an empty string.
	NewOptionOrArgument ContextType = iota
	// NewOption indicates that the last element must be new option, short or
	// long. Returned when it is "-".
	NewOption
	// NewLongOption indicates that the last element must be a new long option.
	// Returned when it is "--".
	NewLongOption
	// LongOption indicates that the last element is a long option, but not its
	// argument. The partial name of the long option is stored in Context.Text.
	LongOption
	// ChainShortOption indicates that a new short option may be chained.
	// Returned when the last element consists of a chain of options that take
	// no arguments.
	ChainShortOption
	// OptionArgument indicates that the last element list must be an argument
	// to an option. The option in question is stored in Context.Option.
	OptionArgument
	// Argument indicates that the last element is an argument. The partial
	// argument is stored in Context.Text.
	Argument
)

func (g *Getopt) findShort(r rune) *Option {
	for _, opt := range g.Options {
		if r == opt.Short {
			return opt
		}
	}
	return nil
}

// parseShort parse short options, without the leading dash. It returns the
// parsed options and whether an argument is still to be seen.
func (g *Getopt) parseShort(s string) ([]*ParsedOption, bool) {
	var opts []*ParsedOption
	var needArg bool
	for i, r := range s {
		opt := g.findShort(r)
		if opt != nil {
			if opt.HasArg == NoArgument {
				opts = append(opts, &ParsedOption{opt, false, ""})
				continue
			} else {
				parsed := &ParsedOption{opt, false, s[i+len(string(r)):]}
				opts = append(opts, parsed)
				needArg = parsed.Argument == "" && opt.HasArg == RequiredArgument
				break
			}
		}
		// Unknown option, treat as taking an optional argument
		parsed := &ParsedOption{
			&Option{r, "", OptionalArgument}, false, s[i+len(string(r)):]}
		opts = append(opts, parsed)
		break
	}
	return opts, needArg
}

// parseLong parse a long option, without the leading dashes. It returns the
// parsed option and whether an argument is still to be seen.
func (g *Getopt) parseLong(s string) (*ParsedOption, bool) {
	eq := strings.IndexRune(s, '=')
	for _, opt := range g.Options {
		if s == opt.Long {
			return &ParsedOption{opt, true, ""}, opt.HasArg == RequiredArgument
		} else if eq != -1 && s[:eq] == opt.Long {
			return &ParsedOption{opt, true, s[eq+1:]}, false
		}
	}
	// Unknown option, treat as taking an optional argument
	if eq == -1 {
		return &ParsedOption{&Option{0, s, OptionalArgument}, true, ""}, false
	}
	return &ParsedOption{&Option{0, s[:eq], OptionalArgument}, true, s[eq+1:]}, false
}

// Parse parses an argument list.
func (g *Getopt) Parse(elems []string) ([]*ParsedOption, []string, *Context) {
	var (
		opts []*ParsedOption
		args []string
		// Non-nil only when the last element was an option with required
		// argument, but the argument has not been seen.
		opt *ParsedOption
		// True if an option terminator has been seen. The criteria of option
		// terminators is determined by the configuration.
		noopt bool
	)
	var elem string
	hasPrefix := func(p string) bool { return strings.HasPrefix(elem, p) }
	for _, elem = range elems[:len(elems)-1] {
		if opt != nil {
			opt.Argument = elem
			opts = append(opts, opt)
			opt = nil
		} else if noopt {
			args = append(args, elem)
		} else if g.Config.HasAll(DoubleDashTerminatesOptions) && elem == "--" {
			noopt = true
		} else if hasPrefix("--") {
			newopt, needArg := g.parseLong(elem[2:])
			if needArg {
				opt = newopt
			} else {
				opts = append(opts, newopt)
			}
		} else if hasPrefix("-") {
			if g.Config.HasAll(LongOnly) {
				newopt, needArg := g.parseLong(elem[1:])
				if needArg {
					opt = newopt
				} else {
					opts = append(opts, newopt)
				}
			} else {
				newopts, needArg := g.parseShort(elem[1:])
				if needArg {
					opts = append(opts, newopts[:len(newopts)-1]...)
					opt = newopts[len(newopts)-1]
				} else {
					opts = append(opts, newopts...)
				}
			}
		} else {
			args = append(args, elem)
			if g.Config.HasAll(FirstArgTerminatesOptions) {
				noopt = true
			}
		}
	}
	elem = elems[len(elems)-1]
	ctx := &Context{}
	if opt != nil {
		opt.Argument = elem
		ctx.Type, ctx.Option = OptionArgument, opt
	} else if noopt {
		ctx.Type, ctx.Text = Argument, elem
	} else if elem == "" {
		ctx.Type = NewOptionOrArgument
	} else if elem == "-" {
		ctx.Type = NewOption
	} else if elem == "--" {
		ctx.Type = NewLongOption
	} else if hasPrefix("--") {
		if strings.IndexRune(elem, '=') == -1 {
			ctx.Type, ctx.Text = LongOption, elem[2:]
		} else {
			newopt, _ := g.parseLong(elem[2:])
			ctx.Type, ctx.Option = OptionArgument, newopt
		}
	} else if hasPrefix("-") {
		if g.Config.HasAll(LongOnly) {
			if strings.IndexRune(elem, '=') == -1 {
				ctx.Type, ctx.Text = LongOption, elem[1:]
			} else {
				newopt, _ := g.parseLong(elem[1:])
				ctx.Type, ctx.Option = OptionArgument, newopt
			}
		} else {
			newopts, _ := g.parseShort(elem[1:])
			if newopts[len(newopts)-1].Option.HasArg == NoArgument {
				opts = append(opts, newopts...)
				ctx.Type = ChainShortOption
			} else {
				opts = append(opts, newopts[:len(newopts)-1]...)
				ctx.Type, ctx.Option = OptionArgument, newopts[len(newopts)-1]
			}
		}
	} else {
		ctx.Type, ctx.Text = Argument, elem
	}
	return opts, args, ctx
}
