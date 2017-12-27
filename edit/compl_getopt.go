package edit

import (
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/getopt"
	"github.com/elves/elvish/parse"
)

func complGetopt(ec *eval.Frame, a []eval.Value, o map[string]eval.Value) {
	var elemsv, optsv, argsv eval.IterableValue
	eval.ScanArgs(a, &elemsv, &optsv, &argsv)
	eval.TakeNoOpt(o)

	var (
		elems    []string
		opts     []*getopt.Option
		args     []eval.CallableValue
		variadic bool
	)
	desc := make(map[*getopt.Option]string)
	// Convert arguments.
	elemsv.Iterate(func(v eval.Value) bool {
		elem, ok := v.(eval.String)
		if !ok {
			throwf("arg should be string, got %s", v.Kind())
		}
		elems = append(elems, string(elem))
		return true
	})
	optsv.Iterate(func(v eval.Value) bool {
		m, ok := v.(eval.MapLike)
		if !ok {
			throwf("opt should be map-like, got %s", v.Kind())
		}
		get := func(ks string) (string, bool) {
			kv := eval.String(ks)
			if !m.HasKey(kv) {
				return "", false
			}
			vv := m.IndexOne(kv)
			if vs, ok := vv.(eval.String); ok {
				return string(vs), true
			} else {
				throwf("%s should be string, got %s", ks, vs.Kind())
				panic("unreachable")
			}
		}

		opt := &getopt.Option{}
		if s, ok := get("short"); ok {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || size != len(s) {
				throwf("short option should be exactly one rune, got %v", parse.Quote(s))
			}
			opt.Short = r
		}
		if s, ok := get("long"); ok {
			opt.Long = s
		}
		if opt.Short == 0 && opt.Long == "" {
			throwf("opt should have at least one of short and long forms")
		}
		if s, ok := get("desc"); ok {
			desc[opt] = s
		}
		opts = append(opts, opt)
		return true
	})
	argsv.Iterate(func(v eval.Value) bool {
		sv, ok := v.(eval.String)
		if ok {
			if string(sv) == "..." {
				variadic = true
				return true
			}
			throwf("string except for ... not allowed as argument handler, got %s", parse.Quote(string(sv)))
		}
		arg, ok := v.(eval.CallableValue)
		if !ok {
			throwf("argument handler should be fn, got %s", v.Kind())
		}
		args = append(args, arg)
		return true
	})
	// TODO Configurable config
	g := getopt.Getopt{opts, getopt.GNUGetoptLong}
	_, parsedArgs, ctx := g.Parse(elems)
	out := ec.OutputChan()

	putShortOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "-" + string(opt.Short)}
		if d, ok := desc[opt]; ok {
			c.displaySuffix = " (" + d + ")"
		}
		out <- c
	}
	putLongOpt := func(opt *getopt.Option) {
		c := &complexCandidate{stem: "--" + string(opt.Long)}
		if d, ok := desc[opt]; ok {
			c.displaySuffix = " (" + d + ")"
		}
		out <- c
	}

	switch ctx.Type {
	case getopt.NewOptionOrArgument, getopt.Argument:
		// Find argument completer
		var argCompl eval.CallableValue
		if len(parsedArgs) < len(args) {
			argCompl = args[len(parsedArgs)]
		} else if variadic {
			argCompl = args[len(args)-1]
		}
		if argCompl != nil {
			rawCands := make(chan rawCandidate)
			defer close(rawCands)
			go func() {
				for rc := range rawCands {
					out <- rc
				}
			}()
			err := callArgCompleter(argCompl, ec.Evaler, []string{ctx.Text}, rawCands)
			maybeThrow(err)
		}
		// TODO Notify that there is no suitable argument completer
	case getopt.NewOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				putShortOpt(opt)
			}
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.NewLongOption:
		for _, opt := range opts {
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.LongOption:
		for _, opt := range opts {
			if strings.HasPrefix(opt.Long, ctx.Text) {
				putLongOpt(opt)
			}
		}
	case getopt.ChainShortOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				// XXX loses chained options
				putShortOpt(opt)
			}
		}
	case getopt.OptionArgument:
	}
}
