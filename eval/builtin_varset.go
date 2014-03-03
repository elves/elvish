package eval

// Implementation of var/set special forms.

import (
	"errors"

	"github.com/xiaq/elvish/parse"
)

type varSetForm struct {
	names  []string
	values []parse.Node
}

var (
	errorBadForm = errors.New("bad form")
)

// parseVarSetForm parses arguments in the var/set special form into a
// varSetForm.
//
// The arguments in the var/set special form must consist of zero or more
// variable factors followed by `=` and then zero or more terms. The number of
// values the terms evaluate to must be equal to the number of names, but
// parseVarSetForm does not attempt to check this.
//
// TODO(xiaq): parseVarSetForm should return more detailed error messsage.
func parseVarSetForm(args *parse.TermListNode) (*varSetForm, error) {
	f := &varSetForm{}
	for i, n := range args.Nodes {
		n := n.(*parse.TermNode)
		if len(n.Nodes) != 1 {
			return nil, errorBadForm
		}
		nf := n.Nodes[0].(*parse.FactorNode)

		var text string
		if m, ok := nf.Node.(*parse.StringNode); ok {
			text = m.Text
		} else {
			return nil, errorBadForm
		}

		if nf.Typ == parse.StringFactor && text == "=" {
			f.values = args.Nodes[i+1:]
			break
		} else if nf.Typ == parse.VariableFactor {
			f.names = append(f.names, text)
		} else {
			return nil, errorBadForm
		}
	}
	return f, nil
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		ev.locals[name] = values[i]
	}

	return ""
}

func var_(ev *Evaluator, args *parse.TermListNode, ports [2]*port) string {
	f, err := parseVarSetForm(args)
	if err != nil {
		return err.Error()
	}
	for _, name := range f.names {
		ev.locals[name] = nil
	}
	if f.values != nil {
		return doSet(ev, f.names, ev.evalTermList(
			&parse.TermListNode{parse.ListNode{0, f.values}}))
	}
	return ""
}

func set(ev *Evaluator, args *parse.TermListNode, ports [2]*port) string {
	f, err := parseVarSetForm(args)
	if err != nil {
		return err.Error()
	}
	if f.values == nil {
		return "not implemented"
	}
	return doSet(ev, f.names, ev.evalTermList(
		&parse.TermListNode{parse.ListNode{0, f.values}}))
}
