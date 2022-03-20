package parse

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
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
//
// The dynamic type of the Node being checked is assumed to be a pointer to a
// struct that embeds the "node" struct.
type ast struct {
	name   string
	fields fs
}

// fs specifies fields of a Node to check. For the value of field $f in the
// Node ("found value"), fs[$f] ("wanted value") is used to check against it.
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
type fs map[string]any

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
		fields := Children(n)
		if len(fields) != 1 {
			return fmt.Errorf("want exactly 1 child, got %d (%s)", len(fields), summary(n))
		}
		n = fields[0]
	}

	ntype := reflect.TypeOf(n).Elem()
	nvalue := reflect.ValueOf(n).Elem()

	for i := 0; i < ntype.NumField(); i++ {
		fieldname := ntype.Field(i).Name
		if !exported(fieldname) {
			// Unexported field
			continue
		}
		got := nvalue.Field(i).Interface()
		want, ok := want.fields[fieldname]
		if ok {
			err := checkField(got, want, "field "+fieldname+" of: "+summary(n))
			if err != nil {
				return err
			}
		} else {
			// Not specified. Check if got is a zero value of its type.
			zero := reflect.Zero(reflect.TypeOf(got)).Interface()
			if !reflect.DeepEqual(got, zero) {
				return fmt.Errorf("want %v, got %v (field %s of: %s)", zero, got, fieldname, summary(n))
			}
		}
	}

	return nil
}

// checkField checks a field against a field specification.
func checkField(got any, want any, ctx string) error {
	// Want nil.
	if want == nil {
		if !reflect.ValueOf(got).IsNil() {
			return fmt.Errorf("want nil, got %v (%s)", got, ctx)
		}
		return nil
	}

	if got, ok := got.(Node); ok {
		// Got a Node.
		return checkNodeInField(got, want)
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

func checkNodeInField(got Node, want any) error {
	switch want := want.(type) {
	case string:
		text := SourceText(got)
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

func exported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
