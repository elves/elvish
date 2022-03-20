package parse

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

const (
	maxL      = 10
	maxR      = 10
	indentInc = 2
)

// Pretty-prints the AST part of a Node to a Writer.
func pprintAST(n Node, w io.Writer) {
	pprintASTRec(n, w, 0, "")
}

type field struct {
	name  string
	tag   reflect.StructTag
	value any
}

var zeroValue reflect.Value

func pprintASTRec(n Node, wr io.Writer, indent int, leading string) {
	nodeType := reflect.TypeOf((*Node)(nil)).Elem()

	var childFields, childrenFields, propertyFields []field

	nt := reflect.TypeOf(n).Elem()
	nv := reflect.ValueOf(n).Elem()

	for i := 0; i < nt.NumField(); i++ {
		f := nt.Field(i)
		if f.Anonymous {
			// embedded node struct, skip
			continue
		}
		ft := f.Type
		fv := nv.Field(i)
		if ft.Kind() == reflect.Slice {
			// list of children
			if ft.Elem().Implements(nodeType) {
				childrenFields = append(childrenFields,
					field{f.Name, f.Tag, fv.Interface()})
				continue
			}
		} else if child, ok := fv.Interface().(Node); ok {
			// a child node
			if reflect.Indirect(fv) != zeroValue {
				childFields = append(childFields,
					field{f.Name, f.Tag, child})
			}
			continue
		}
		// a property
		propertyFields = append(propertyFields,
			field{f.Name, f.Tag, fv.Interface()})
	}

	// has only one child and nothing more : coalesce
	if len(Children(n)) == 1 &&
		SourceText(Children(n)[0]) == SourceText(n) {
		pprintASTRec(Children(n)[0], wr, indent, leading+nt.Name()+"/")
		return
	}
	// print heading
	//b := n.n()
	//fmt.Fprintf(wr, "%*s%s%s %s %d-%d", indent, "",
	//	wr.leading, nt.Name(), compactQuote(b.source(src)), b.begin, b.end)
	fmt.Fprintf(wr, "%*s%s%s", indent, "", leading, nt.Name())
	// print properties
	for _, pf := range propertyFields {
		fmtstring := pf.tag.Get("fmt")
		if len(fmtstring) > 0 {
			fmt.Fprintf(wr, " %s="+fmtstring, pf.name, pf.value)
		} else {
			value := pf.value
			if s, ok := value.(string); ok {
				value = compactQuote(s)
			}
			fmt.Fprintf(wr, " %s=%v", pf.name, value)
		}
	}
	fmt.Fprint(wr, "\n")
	// print lone children recursively
	for _, chf := range childFields {
		// TODO the name is omitted
		pprintASTRec(chf.value.(Node), wr, indent+indentInc, "")
	}
	// print children list recursively
	for _, chf := range childrenFields {
		children := reflect.ValueOf(chf.value)
		if children.Len() == 0 {
			continue
		}
		// fmt.Fprintf(wr, "%*s.%s:\n", indent, "", chf.name)
		for i := 0; i < children.Len(); i++ {
			n := children.Index(i).Interface().(Node)
			pprintASTRec(n, wr, indent+indentInc, "")
		}
	}
}

// Pretty-prints the parse tree part of a Node to a Writer.
func pprintParseTree(n Node, w io.Writer) {
	pprintParseTreeRec(n, w, 0)
}

func pprintParseTreeRec(n Node, wr io.Writer, indent int) {
	leading := ""
	for len(Children(n)) == 1 {
		leading += reflect.TypeOf(n).Elem().Name() + "/"
		n = Children(n)[0]
	}
	fmt.Fprintf(wr, "%*s%s%s\n", indent, "", leading, summary(n))
	for _, ch := range Children(n) {
		pprintParseTreeRec(ch, wr, indent+indentInc)
	}
}

func summary(n Node) string {
	return fmt.Sprintf("%s %s %d-%d", reflect.TypeOf(n).Elem().Name(),
		compactQuote(SourceText(n)), n.Range().From, n.Range().To)
}

func compactQuote(text string) string {
	if len(text) > maxL+maxR+3 {
		text = text[0:maxL] + "..." + text[len(text)-maxR:]
	}
	return strconv.Quote(text)
}
