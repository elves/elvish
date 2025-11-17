// Structoffunc generates "struct of func" definitions for interfaces.
//
// See [etk.RegisterStructOfFuncForInterface] for the "struct of func" pattern.
//
// This program generates a "struct of func" type for each interface.
// The CLI flags mimic that of stringer:
//
//	structoffunc -type=A,B,C -output=zstructoffunc.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	"src.elv.sh/pkg/strutil"
)

type imports struct {
	fileNameToPath map[string]string
	outNameToPath  map[string]string
	outPathToName  map[string]string
}

var (
	typeFlag   = flag.String("type", "", "comma-separated list of interface type names; must be set")
	outputFlag = flag.String("output", "zstructoffunc.go", "output file name; default zstructoffunc.go")
)

func main() {
	flag.Parse()
	if *typeFlag == "" {
		flag.Usage()
		os.Exit(1)
	}
	typesSet := map[string]bool{}
	for _, t := range strings.Split(*typeFlag, ",") {
		typesSet[t] = true
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil || packages.PrintErrors(pkgs) > 0 || len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "Failed to load package: %v\n", err)
		os.Exit(1)
	}
	pkg := pkgs[0]

	outImportsNameToPath := map[string]string{"etk": "src.elv.sh/pkg/etk"}
	outImportsPathToName := map[string]string{"src.elv.sh/pkg/etk": "etk"}
	var outDefWriter strings.Builder

	for _, file := range pkg.Syntax {
		imports := imports{map[string]string{}, outImportsNameToPath, outImportsPathToName}
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch genDecl.Tok {
			case token.IMPORT:
				for _, spec := range genDecl.Specs {
					importSpec := spec.(*ast.ImportSpec)
					pkgPath, _ := strconv.Unquote(importSpec.Path.Value)
					var pkgName string
					if importSpec.Name != nil {
						pkgName = importSpec.Name.Name
					} else {
						pkgName = path.Base(pkgPath)
					}
					imports.fileNameToPath[pkgName] = pkgPath
				}
			case token.TYPE:
				for _, spec := range genDecl.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					iface, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok || !typesSet[typeSpec.Name.Name] {
						continue
					}
					outDefWriter.WriteString(genStructOfFunc(typeSpec.Name.Name, iface, imports))
				}
			}
		}
	}

	err = writeOutput(*outputFlag, pkg.Name, outImportsNameToPath, outDefWriter.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}
}

func genStructOfFunc(ifaceName string, iface *ast.InterfaceType, imports imports) string {
	var decl strings.Builder
	var impl strings.Builder
	structName := "sof" + ifaceName
	fmt.Fprintf(&decl, "var _ %s = %s{}\n", ifaceName, structName)
	fmt.Fprintf(&decl, "var _ = etk.RegisterStructOfFuncForInterface[%s, %s]()\n\n",
		structName, ifaceName)
	fmt.Fprintf(&decl, "type %s struct {\n", structName)
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			// TODO: Support embedded interfaces.
			continue
		}
		methodName := method.Names[0].Name
		funcType := method.Type.(*ast.FuncType)
		fmt.Fprintf(&decl, "\t %sImpl %s `elvish:%q`\n",
			methodName, genType(funcType, imports),
			strutil.CamelToDashed(methodName))

		if impl.Len() > 0 {
			impl.WriteString("\n")
		}
		impl.WriteString(genMethodImpl(structName, methodName, funcType, imports))
	}
	fmt.Fprintf(&decl, "}\n\n")

	return decl.String() + impl.String() + "\n"
}

func genMethodImpl(structName, methodName string, funcType *ast.FuncType, imports imports) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "func (sof %s) %s(%s) %s {\n\t",
		structName, methodName,
		genParamList(funcType.Params, imports), genReturnType(funcType.Results, imports))
	if funcType.Results != nil {
		sb.WriteString("return ")
	}
	fmt.Fprintf(&sb, "sof.%sImpl(%s)", methodName, genArgList(funcType.Params))
	sb.WriteString("\n}\n")
	return sb.String()
}

func genReturnType(results *ast.FieldList, imports imports) string {
	if results == nil {
		return ""
	} else if len(results.List) == 1 {
		return genType(results.List[0].Type, imports)
	} else {
		return "(" + genTypeList(results, imports) + ")"
	}
}

func genParamList(params *ast.FieldList, imports imports) string {
	var parts []string
	for i, param := range params.List {
		name := fmt.Sprintf("a%d", i+1)
		parts = append(parts, fmt.Sprintf("%s %s", name, genType(param.Type, imports)))
	}
	return strings.Join(parts, ", ")
}

func genArgList(params *ast.FieldList) string {
	var parts []string
	for i, param := range params.List {
		name := fmt.Sprintf("a%d", i+1)
		if _, isEllipsis := param.Type.(*ast.Ellipsis); isEllipsis {
			name += "..."
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func genTypeList(typeList *ast.FieldList, imports imports) string {
	var parts []string
	for _, field := range typeList.List {
		parts = append(parts, genType(field.Type, imports))
	}
	return strings.Join(parts, ", ")
}

func genType(expr ast.Expr, imports imports) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", genType(t.Elt, imports))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", genModuleName(t.X.(*ast.Ident).Name, imports), t.Sel.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", genType(t.X, imports))
	case *ast.ArrayType:
		if t.Len != nil {
			// TODO: Handle more general length expressions.
			return fmt.Sprintf("[%s]%s", t.Len.(*ast.BasicLit).Value, genType(t.Elt, imports))
		} else {
			return fmt.Sprintf("[]%s", genType(t.Elt, imports))
		}
	case *ast.StructType:
		var sb strings.Builder
		sb.WriteString("struct {\n")
		for _, field := range t.Fields.List {
			fmt.Fprintf(&sb, "\t%s\n", genField(field, imports))
		}
		sb.WriteString("}")
		return sb.String()
	case *ast.FuncType:
		var sb strings.Builder
		fmt.Fprintf(&sb, "func(%s) %s",
			genTypeList(t.Params, imports), genReturnType(t.Results, imports))
		return sb.String()
	case *ast.InterfaceType:
		var sb strings.Builder
		sb.WriteString("interface {\n")
		for _, method := range t.Methods.List {
			t := method.Type.(*ast.FuncType)
			fmt.Fprintf(&sb, "\t%s(%s) %s\n",
				method.Names[0].Name,
				genTypeList(t.Params, imports), genReturnType(t.Results, imports))
		}
		sb.WriteString("}")
		return sb.String()
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", genType(t.Key, imports),
			genType(t.Value, imports))
	case *ast.ChanType:
		chanToken := "chan"
		switch t.Dir {
		case ast.SEND:
			chanToken = "chan<-"
		case ast.RECV:
			chanToken = "<-chan"
		}
		return fmt.Sprintf("%s %s", chanToken, genType(t.Value, imports))
	default:
		panic("unsupported type expression: " + fmt.Sprintf("%T", expr))
	}
}

func genField(field *ast.Field, imports imports) string {
	names := make([]string, len(field.Names))
	for i, name := range field.Names {
		names[i] = name.Name
	}
	return fmt.Sprintf("%s %s",
		strings.Join(names, ", "), genType(field.Type, imports))
}

func genModuleName(name string, imports imports) string {
	path, ok := imports.fileNameToPath[name]
	if !ok {
		// TODO: Don't panic
		panic(fmt.Sprintf("unknown import name %s, import map is %s", name, imports.fileNameToPath))
	}
	if outName, ok := imports.outPathToName[path]; ok {
		return outName
	}
	// Create a new import in the output file.
	outName := name
	suffix := 2
	for {
		if _, exists := imports.outNameToPath[outName]; !exists {
			imports.outNameToPath[outName] = path
			imports.outPathToName[path] = outName
			return outName
		}
		outName = fmt.Sprintf("%s%d", name, suffix)
		suffix++
	}
}

func writeOutput(filename, pkgname string, importsNameToPath map[string]string, defs string) error {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated by structoffunc; DO NOT EDIT.")
	fmt.Fprintf(&buf, "//go:generate go run src.elv.sh/cmd/structoffunc -type=%s -output=%s\n",
		*typeFlag, *outputFlag)
	fmt.Fprintf(&buf, "package %s\n\n", pkgname)
	fmt.Fprintln(&buf, "import (")
	for name, pkgPath := range importsNameToPath {
		fmt.Fprintf(&buf, "\t%s \"%s\"\n", name, pkgPath)
	}
	fmt.Fprintf(&buf, ")\n\n%s", defs)

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	return os.WriteFile(filename, formatted, 0644)
}
