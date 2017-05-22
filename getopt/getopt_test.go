package getopt

import (
	"reflect"
	"testing"
)

var options = []*Option{
	{'a', "all", NoArgument},
	{'o', "option", RequiredArgument},
	{'n', "number", OptionalArgument},
}

var cases = []struct {
	config   Config
	elems    []string
	wantOpts []*ParsedOption
	wantArgs []string
	wantCtx  *Context
}{
	// NoArgument, short option.
	{0, []string{"-a", ""},
		[]*ParsedOption{{options[0], false, ""}},
		nil, &Context{Type: NewOptionOrArgument}},
	// NoArgument, long option.
	{0, []string{"--all", ""},
		[]*ParsedOption{{options[0], true, ""}},
		nil, &Context{Type: NewOptionOrArgument}},

	// RequiredArgument, argument following the option directly
	{0, []string{"-oname=elvish", ""},
		[]*ParsedOption{{options[1], false, "name=elvish"}},
		nil, &Context{Type: NewOptionOrArgument}},
	// RequiredArgument, argument in next element
	{0, []string{"-o", "name=elvish", ""},
		[]*ParsedOption{{options[1], false, "name=elvish"}},
		nil, &Context{Type: NewOptionOrArgument}},
	// RequiredArgument, long option, argument following the option directly
	{0, []string{"--option=name=elvish", ""},
		[]*ParsedOption{{options[1], true, "name=elvish"}},
		nil, &Context{Type: NewOptionOrArgument}},
	// RequiredArgument, long option, argument in next element
	{0, []string{"--option", "name=elvish", ""},
		[]*ParsedOption{{options[1], true, "name=elvish"}},
		nil, &Context{Type: NewOptionOrArgument}},

	// OptionalArgument, with argument
	{0, []string{"-n1", ""},
		[]*ParsedOption{{options[2], false, "1"}},
		nil, &Context{Type: NewOptionOrArgument}},
	// OptionalArgument, without argument
	{0, []string{"-n", ""},
		[]*ParsedOption{{options[2], false, ""}},
		nil, &Context{Type: NewOptionOrArgument}},

	// DoubleDashTerminatesOptions
	{DoubleDashTerminatesOptions, []string{"-a", "--", "-o", ""},
		[]*ParsedOption{{options[0], false, ""}},
		[]string{"-o"}, &Context{Type: Argument}},
	// FirstArgTerminatesOptions
	{FirstArgTerminatesOptions, []string{"-a", "x", "-o", ""},
		[]*ParsedOption{{options[0], false, ""}},
		[]string{"x", "-o"}, &Context{Type: Argument}},
	// LongOnly
	{LongOnly, []string{"-all", ""},
		[]*ParsedOption{{options[0], true, ""}},
		nil, &Context{Type: NewOptionOrArgument}},

	// NewOption
	{0, []string{"-"}, nil, nil, &Context{Type: NewOption}},
	// NewLongOption
	{0, []string{"--"}, nil, nil, &Context{Type: NewLongOption}},
	// LongOption
	{0, []string{"--all"}, nil, nil,
		&Context{Type: LongOption, Text: "all"}},
	// LongOption, LongOnly
	{LongOnly, []string{"-all"}, nil, nil,
		&Context{Type: LongOption, Text: "all"}},
	// ChainShortOption
	{0, []string{"-a"},
		[]*ParsedOption{{options[0], false, ""}}, nil,
		&Context{Type: ChainShortOption}},
	// OptionArgument, short option, same element
	{0, []string{"-o"}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], false, ""}}},
	// OptionArgument, short option, separate element
	{0, []string{"-o", ""}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], false, ""}}},
	// OptionArgument, long option, same element
	{0, []string{"--option="}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], true, ""}}},
	// OptionArgument, long option, separate element
	{0, []string{"--option", ""}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], true, ""}}},
	// OptionArgument, long only, same element
	{LongOnly, []string{"-option="}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], true, ""}}},
	// OptionArgument, long only, separate element
	{LongOnly, []string{"-option", ""}, nil, nil,
		&Context{Type: OptionArgument,
			Option: &ParsedOption{options[1], true, ""}}},
	// Argument
	{0, []string{"x"}, nil, nil,
		&Context{Type: Argument, Text: "x"}},

	// Unknown short option, same element
	{0, []string{"-x"}, nil, nil,
		&Context{
			Type: OptionArgument,
			Option: &ParsedOption{
				&Option{'x', "", OptionalArgument}, false, ""}}},
	// Unknown short option, separate element
	{0, []string{"-x", ""},
		[]*ParsedOption{{
			&Option{'x', "", OptionalArgument}, false, ""}},
		nil,
		&Context{Type: NewOptionOrArgument}},

	// Unknown long option
	{0, []string{"--unknown", ""},
		[]*ParsedOption{{
			&Option{0, "unknown", OptionalArgument}, true, ""}},
		nil,
		&Context{Type: NewOptionOrArgument}},
	// Unknown long option, with argument
	{0, []string{"--unknown=value", ""},
		[]*ParsedOption{{
			&Option{0, "unknown", OptionalArgument}, true, "value"}},
		nil,
		&Context{Type: NewOptionOrArgument}},

	// Unknown long option, LongOnly
	{LongOnly, []string{"-unknown", ""},
		[]*ParsedOption{{
			&Option{0, "unknown", OptionalArgument}, true, ""}},
		nil,
		&Context{Type: NewOptionOrArgument}},
	// Unknown long option, with argument
	{LongOnly, []string{"-unknown=value", ""},
		[]*ParsedOption{{
			&Option{0, "unknown", OptionalArgument}, true, "value"}},
		nil,
		&Context{Type: NewOptionOrArgument}},
}

func TestGetopt(t *testing.T) {
	for _, tc := range cases {
		g := &Getopt{options, tc.config}
		opts, args, ctx := g.Parse(tc.elems)
		shouldEqual := func(name string, got, want interface{}) {
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Parse(%#v) (config = %v)\ngot %s = %v, want %v",
					tc.elems, tc.config, name, got, want)
			}
		}
		shouldEqual("opts", opts, tc.wantOpts)
		shouldEqual("args", args, tc.wantArgs)
		shouldEqual("ctx", ctx, tc.wantCtx)
	}
}
