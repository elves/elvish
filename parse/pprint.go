package parse

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

const (
	maxL      int = 10
	maxR          = 10
	indentInc     = 2
)

// PPrintAST pretty-prints the AST part of a Node to a Writer.
func PPrintAST(n Node, w io.Writer) {
	pprintAST(n, w, 0, "")
}

type field struct {
	name  string
	tag   reflect.StructTag
	value interface{}
}

var zeroValue reflect.Value

func pprintAST(n Node, wr io.Writer, indent int, leading string) {
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
	if len(n.Children()) == 1 &&
		n.Children()[0].SourceText() == n.SourceText() {
		pprintAST(n.Children()[0], wr, indent, leading+nt.Name()+"/")
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
		pprintAST(chf.value.(Node), wr, indent+indentInc, "")
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
			pprintAST(n, wr, indent+indentInc, "")
		}
	}
}

// PPrintParseTree pretty-prints the parse tree part of a Node to a Writer.
func PPrintParseTree(n Node, w io.Writer) {
	pprintParseTree(n, w, 0)
}

func pprintParseTree(n Node, wr io.Writer, indent int) {
	leading := ""
	for len(n.Children()) == 1 {
		leading += reflect.TypeOf(n).Elem().Name() + "/"
		n = n.Children()[0]
	}
	fmt.Fprintf(wr, "%*s%s%s\n", indent, "", leading, summary(n))
	for _, ch := range n.Children() {
		pprintParseTree(ch, wr, indent+indentInc)
	}
}

func summary(n Node) string {
	return fmt.Sprintf("%s %s %d-%d", reflect.TypeOf(n).Elem().Name(),
		compactQuote(n.SourceText()), n.Range().From, n.Range().To)
}

func compactQuote(text string) string {
	if len(text) > maxL+maxR+3 {
		text = text[0:maxL] + "..." + text[len(text)-maxR:]
	}
	return strconv.Quote(text)
}
