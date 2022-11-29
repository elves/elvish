package getopt

import (
	"errors"
	"reflect"
	"testing"

	"src.elv.sh/pkg/errutil"
)

var (
	vSpec = &OptionSpec{'v', "verbose", NoArgument}
	nSpec = &OptionSpec{'n', "dry-run", NoArgument}
	fSpec = &OptionSpec{'f', "file", RequiredArgument}
	iSpec = &OptionSpec{'i', "in-place", OptionalArgument}
	specs = []*OptionSpec{vSpec, nSpec, fSpec, iSpec}
)

var parseTests = []struct {
	name     string
	cfg      Config
	args     []string
	wantOpts []*Option
	wantArgs []string
	wantErr  error
}{
	{
		name:     "short option",
		args:     []string{"-v"},
		wantOpts: []*Option{{Spec: vSpec}},
	},
	{
		name:     "short option with required argument",
		args:     []string{"-fname"},
		wantOpts: []*Option{{Spec: fSpec, Argument: "name"}},
	},
	{
		name:     "short option with required argument in separate argument",
		args:     []string{"-f", "name"},
		wantOpts: []*Option{{Spec: fSpec, Argument: "name"}},
	},
	{
		name:     "short option with optional argument",
		args:     []string{"-i.bak"},
		wantOpts: []*Option{{Spec: iSpec, Argument: ".bak"}},
	},
	{
		name:     "short option with optional argument omitted",
		args:     []string{"-i", ".bak"},
		wantOpts: []*Option{{Spec: iSpec}},
		wantArgs: []string{".bak"},
	},
	{
		name:     "short option chaining",
		args:     []string{"-vn"},
		wantOpts: []*Option{{Spec: vSpec}, {Spec: nSpec}},
	},
	{
		name:     "short option chaining with argument",
		args:     []string{"-vfname"},
		wantOpts: []*Option{{Spec: vSpec}, {Spec: fSpec, Argument: "name"}},
	},
	{
		name:     "short option chaining with argument in separate argument",
		args:     []string{"-vf", "name"},
		wantOpts: []*Option{{Spec: vSpec}, {Spec: fSpec, Argument: "name"}},
	},

	{
		name:     "long option",
		args:     []string{"--verbose"},
		wantOpts: []*Option{{Spec: vSpec, Long: true}},
	},
	{
		name:     "long option with required argument",
		args:     []string{"--file=name"},
		wantOpts: []*Option{{Spec: fSpec, Long: true, Argument: "name"}},
	},
	{
		name:     "long option with required argument in separate argument",
		args:     []string{"--file", "name"},
		wantOpts: []*Option{{Spec: fSpec, Long: true, Argument: "name"}},
	},
	{
		name:     "long option with optional argument",
		args:     []string{"--in-place=.bak"},
		wantOpts: []*Option{{Spec: iSpec, Long: true, Argument: ".bak"}},
	},
	{
		name:     "long option with optional argument omitted",
		args:     []string{"--in-place", ".bak"},
		wantOpts: []*Option{{Spec: iSpec, Long: true}},
		wantArgs: []string{".bak"},
	},

	{
		name:     "long option, LongOnly mode",
		args:     []string{"-verbose"},
		cfg:      LongOnly,
		wantOpts: []*Option{{Spec: vSpec, Long: true}},
	},
	{
		name:     "long option with required argument, LongOnly mode",
		args:     []string{"-file", "name"},
		cfg:      LongOnly,
		wantOpts: []*Option{{Spec: fSpec, Long: true, Argument: "name"}},
	},

	{
		name:     "StopAfterDoubleDash off",
		args:     []string{"-v", "--", "-n"},
		wantOpts: []*Option{{Spec: vSpec}, {Spec: nSpec}},
		wantArgs: []string{"--"},
	},
	{
		name:     "StopAfterDoubleDash on",
		args:     []string{"-v", "--", "-n"},
		cfg:      StopAfterDoubleDash,
		wantOpts: []*Option{{Spec: vSpec}},
		wantArgs: []string{"-n"},
	},

	{
		name:     "StopBeforeFirstNonOption off",
		args:     []string{"-v", "foo", "-n"},
		wantOpts: []*Option{{Spec: vSpec}, {Spec: nSpec}},
		wantArgs: []string{"foo"},
	},
	{
		name:     "StopBeforeFirstNonOption on",
		args:     []string{"-v", "foo", "-n"},
		cfg:      StopBeforeFirstNonOption,
		wantOpts: []*Option{{Spec: vSpec}},
		wantArgs: []string{"foo", "-n"},
	},

	{
		name:     "single dash is not an option",
		args:     []string{"-"},
		wantArgs: []string{"-"},
	},
	{
		name:     "single dash is not an option, LongOnly mode",
		args:     []string{"-"},
		cfg:      LongOnly,
		wantArgs: []string{"-"},
	},

	{
		name:    "short option with required argument missing",
		args:    []string{"-f"},
		wantErr: errors.New("missing argument for -f"),
	},
	{
		name:    "long option with required argument missing",
		args:    []string{"--file"},
		wantErr: errors.New("missing argument for --file"),
	},
	{
		name: "unknown short option",
		args: []string{"-b"},
		wantOpts: []*Option{
			{Spec: &OptionSpec{Short: 'b', Arity: OptionalArgument}, Unknown: true}},
		wantErr: errors.New("unknown option -b"),
	},
	{
		name: "unknown short option with argument",
		args: []string{"-bfoo"},
		wantOpts: []*Option{
			{Spec: &OptionSpec{Short: 'b', Arity: OptionalArgument}, Unknown: true, Argument: "foo"}},
		wantErr: errors.New("unknown option -b"),
	},
	{
		name: "unknown long option",
		args: []string{"--bad"},
		wantOpts: []*Option{
			{Spec: &OptionSpec{Long: "bad", Arity: OptionalArgument}, Long: true, Unknown: true}},
		wantErr: errors.New("unknown option --bad"),
	},
	{
		name: "unknown long option with argument",
		args: []string{"--bad=foo"},
		wantOpts: []*Option{
			{Spec: &OptionSpec{Long: "bad", Arity: OptionalArgument}, Long: true, Unknown: true, Argument: "foo"}},
		wantErr: errors.New("unknown option --bad"),
	},
	{
		name: "multiple errors",
		args: []string{"-b", "-f"},
		wantOpts: []*Option{
			{Spec: &OptionSpec{Short: 'b', Arity: OptionalArgument}, Unknown: true}},
		wantErr: errutil.Multi(
			errors.New("missing argument for -f"), errors.New("unknown option -b")),
	},
}

func TestParse(t *testing.T) {
	for _, tc := range parseTests {
		t.Run(tc.name, func(t *testing.T) {
			opts, args, err := Parse(tc.args, specs, tc.cfg)
			check := func(name string, got, want any) {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Parse(%#v) (config = %v)\ngot %s = %v, want %v",
						tc.args, tc.cfg, name, got, want)
				}
			}
			check("opts", opts, tc.wantOpts)
			check("args", args, tc.wantArgs)
			check("err", err, tc.wantErr)
		})
	}
}

var completeTests = []struct {
	name     string
	cfg      Config
	args     []string
	wantOpts []*Option
	wantArgs []string
	wantCtx  Context
}{
	{
		name:    "NewOptionOrArgument",
		args:    []string{""},
		wantCtx: Context{Type: OptionOrArgument},
	},
	{
		name:    "NewOption",
		args:    []string{"-"},
		wantCtx: Context{Type: AnyOption},
	},
	{
		name:    "LongOption",
		args:    []string{"--f"},
		wantCtx: Context{Type: LongOption, Text: "f"},
	},
	{
		name:    "LongOption with LongOnly",
		args:    []string{"-f"},
		cfg:     LongOnly,
		wantCtx: Context{Type: LongOption, Text: "f"},
	},
	{
		name:     "ChainShortOption",
		args:     []string{"-v"},
		wantOpts: []*Option{{Spec: vSpec}},
		wantCtx:  Context{Type: ChainShortOption},
	},
	{
		name: "OptionArgument of short option, separate argument",
		args: []string{"-f", "foo"},
		wantCtx: Context{
			Type:   OptionArgument,
			Option: &Option{Spec: fSpec, Argument: "foo"}},
	},
	{
		name: "OptionArgument of short option, same argument",
		args: []string{"-ffoo"},
		wantCtx: Context{
			Type:   OptionArgument,
			Option: &Option{Spec: fSpec, Argument: "foo"}},
	},
	{
		name: "OptionArgument of long option, separate argument",
		args: []string{"--file", "foo"},
		wantCtx: Context{
			Type:   OptionArgument,
			Option: &Option{Spec: fSpec, Long: true, Argument: "foo"}},
	},
	{
		name: "OptionArgument of long option, same argument",
		args: []string{"--file=foo"},
		wantCtx: Context{
			Type:   OptionArgument,
			Option: &Option{Spec: fSpec, Long: true, Argument: "foo"}},
	},
	{
		name: "OptionArgument of long option with LongOnly, same argument",
		args: []string{"-file=foo"},
		cfg:  LongOnly,
		wantCtx: Context{
			Type:   OptionArgument,
			Option: &Option{Spec: fSpec, Long: true, Argument: "foo"}},
	},
	{
		name:    "Argument",
		args:    []string{"foo"},
		wantCtx: Context{Type: Argument, Text: "foo"},
	},
	{
		name:    "Argument after --",
		args:    []string{"--", "foo"},
		cfg:     StopAfterDoubleDash,
		wantCtx: Context{Type: Argument, Text: "foo"},
	},
	{
		name:     "Argument after first non-option argument",
		args:     []string{"bar", "foo"},
		cfg:      StopBeforeFirstNonOption,
		wantArgs: []string{"bar"},
		wantCtx:  Context{Type: Argument, Text: "foo"},
	},
}

func TestComplete(t *testing.T) {
	for _, tc := range completeTests {
		t.Run(tc.name, func(t *testing.T) {
			opts, args, ctx := Complete(tc.args, specs, tc.cfg)
			check := func(name string, got, want any) {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Parse(%#v) (config = %v)\ngot %s = %v, want %v",
						tc.args, tc.cfg, name, got, want)
				}
			}
			check("opts", opts, tc.wantOpts)
			check("args", args, tc.wantArgs)
			check("ctx", ctx, tc.wantCtx)
		})
	}
}
