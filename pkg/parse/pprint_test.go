package parse

import (
	"strings"
	"testing"

	"src.elv.sh/pkg/tt"
)

var n = mustParse("ls $x[0]$y[1];echo done >/redir-dest")

var pprintASTTests = tt.Table{
	tt.Args(n).Rets(
		`Chunk
  Pipeline/Form
    Compound/Indexing/Primary ExprCtx=CmdExpr Type=Bareword LegacyLambda=false Value="ls"
    Compound ExprCtx=NormalExpr
      Indexing ExprCtx=NormalExpr
        Primary ExprCtx=NormalExpr Type=Variable LegacyLambda=false Value="x"
        Array/Compound/Indexing/Primary ExprCtx=NormalExpr Type=Bareword LegacyLambda=false Value="0"
      Indexing ExprCtx=NormalExpr
        Primary ExprCtx=NormalExpr Type=Variable LegacyLambda=false Value="y"
        Array/Compound/Indexing/Primary ExprCtx=NormalExpr Type=Bareword LegacyLambda=false Value="1"
  Pipeline/Form
    Compound/Indexing/Primary ExprCtx=CmdExpr Type=Bareword LegacyLambda=false Value="echo"
    Compound/Indexing/Primary ExprCtx=NormalExpr Type=Bareword LegacyLambda=false Value="done"
    Redir Mode=Write RightIsFd=false
      Compound/Indexing/Primary ExprCtx=NormalExpr Type=Bareword LegacyLambda=false Value="/redir-dest"
`),
}

func TestPPrintAST(t *testing.T) {
	tt.Test(t, tt.Fn("PPrintAST (to string)", func(n Node) string {
		var b strings.Builder
		pprintAST(n, &b)
		return b.String()
	}), pprintASTTests)
}

var pprintParseTreeTests = tt.Table{
	tt.Args(n).Rets(
		`Chunk "ls $x[0]$y...redir-dest" 0-36
  Pipeline/Form "ls $x[0]$y[1]" 0-13
    Compound/Indexing/Primary "ls" 0-2
    Sep " " 2-3
    Compound "$x[0]$y[1]" 3-13
      Indexing "$x[0]" 3-8
        Primary "$x" 3-5
        Sep "[" 5-6
        Array/Compound/Indexing/Primary "0" 6-7
        Sep "]" 7-8
      Indexing "$y[1]" 8-13
        Primary "$y" 8-10
        Sep "[" 10-11
        Array/Compound/Indexing/Primary "1" 11-12
        Sep "]" 12-13
  Sep ";" 13-14
  Pipeline/Form "echo done >/redir-dest" 14-36
    Compound/Indexing/Primary "echo" 14-18
    Sep " " 18-19
    Compound/Indexing/Primary "done" 19-23
    Sep " " 23-24
    Redir ">/redir-dest" 24-36
      Sep ">" 24-25
      Compound/Indexing/Primary "/redir-dest" 25-36
`),
}

func TestPPrintParseTree(t *testing.T) {
	tt.Test(t, tt.Fn("PPrintParseTree (to string)", func(n Node) string {
		var b strings.Builder
		pprintParseTree(n, &b)
		return b.String()
	}), pprintParseTreeTests)
}

func mustParse(src string) Node {
	tree, err := Parse(SourceForTest(src), Config{})
	if err != nil {
		panic(err)
	}
	return tree.Root
}
