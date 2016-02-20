package parse

import (
	"fmt"
	"reflect"
	"strings"
)

// AST checking utilities. Used in test cases.

// ast is an AST specification. The name part identifies the type of the Node;
// for instance, "Chunk" specifies a Chunk. The fields part is specifies children
// to check; see document of fs.
//
// When a Node contains exactly one child, It can be coalesced with its child
// by adding "/ChildName" in the name part. For instance, "Chunk/Pipeline"
// specifies a Chunk that contains exactly one Pipeline. In this case, the
// fields part specified the children of the Pipeline instead of the Chunk
// (which has no additional interesting fields anyway). Multi-level coalescence
// like "Chunk/Pipeline/Form" is also allowed.
type ast struct {
	name   string
	fields fs
}

// fs specifies fields of a Node to check. For each key/value pair in fs, the
// value ("wanted value") checks a field in the Node ("found value") using the
// following algorithm:
//
// If the key is "text", the SourceText of the Node is checked. It doesn't
// involve a found value.
//
// If the wanted value is nil, the found value is checked against nil.
//
// If the found value implements Node, then the wanted value must be either an
// ast, where the checking algorithm of ast applies, or a string, where the
// source text of the found value is checked.
//
// If the found value is a slice whose elements implement Node, then the wanted
// value must be a slice where checking is then done recursively.
//
// If the found value satisfied none of the above conditions, it is checked
// against the wanted value using reflect.DeepEqual.
type fs map[string]interface{}

// checkAST checks an AST against a specification.
func checkAST(n Node, want ast) error {
	wantnames := strings.Split(want.name, "/")
	// Check coalesced levels
	for i, wantname := range wantnames {
		name := reflect.TypeOf(n).Elem().Name()
		if wantname != name {
			return fmt.Errorf("want %s, got %s (%s)", wantname, name, summary(n))
		}
		if i == len(wantnames)-1 {
			break
		}
		fields := n.Children()
		if len(fields) != 1 {
			return fmt.Errorf("want exactly 1 child, got %d (%s)", len(fields), summary(n))
		}
		n = fields[0]
	}

	if want.fields == nil && len(n.Children()) != 0 {
		return fmt.Errorf("want leaf, got inner node (%s)", summary(n))
	}
	nv := reflect.ValueOf(n).Elem()

	// TODO: Check fields present in n but not in want
	for fieldname, wantfield := range want.fields {
		if fieldname == "text" {
			if n.SourceText() != wantfield.(string) {
				return fmt.Errorf("want %q, got %q (%s)", wantfield, n.SourceText())
			}
		} else {
			fv := nv.FieldByName(fieldname)
			err := checkField(fv.Interface(), wantfield, summary(n))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var nodeType = reflect.TypeOf((*Node)(nil)).Elem()

// checkField checks a field against a field specification.
func checkField(got interface{}, want interface{}, ctx string) error {
	// Want nil.
	if want == nil {
		if !reflect.ValueOf(got).IsNil() {
			return fmt.Errorf("want nil, got %v (%s)", got, ctx)
		}
		return nil
	}

	if got, ok := got.(Node); ok {
		// Got a Node.
		return checkNodeInField(got.(Node), want)
	}
	tgot := reflect.TypeOf(got)
	if tgot.Kind() == reflect.Slice && tgot.Elem().Implements(nodeType) {
		// Got a slice of Nodes.
		vgot := reflect.ValueOf(got)
		vwant := reflect.ValueOf(want)
		if vgot.Len() != vwant.Len() {
			return fmt.Errorf("want %d, got %d (%s)", vwant.Len(), vgot.Len(), ctx)
		}
		for i := 0; i < vgot.Len(); i++ {
			err := checkNodeInField(vgot.Index(i).Interface().(Node),
				vwant.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("want %v, got %v (%s)", want, got, ctx)
	}
	return nil
}

func checkNodeInField(got Node, want interface{}) error {
	switch want := want.(type) {
	case string:
		text := got.SourceText()
		if want != text {
			return fmt.Errorf("want %q, got %q (%s)", want, text, summary(got))
		}
		return nil
	case ast:
		return checkAST(got, want)
	default:
		panic(fmt.Sprintf("bad want type %T (%s)", want, summary(got)))
	}
}
