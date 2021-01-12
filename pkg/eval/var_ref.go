package eval

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/prog"
)

// This file implements variable resolution. Elvish has fully static lexical
// scopes, so variable resolution involves some work in the compilation phase as
// well.
//
// During compilation, a qualified variable name (whether in lvalue, like "x
// = foo", or in variable use, like "$x") is searched in compiler's staticNs
// tables to determine which scope they belong to, as well as their indicies in
// that scope. This step is just called "resolve" in the code, and it stores
// information in a varRef struct.
//
// During evaluation, the varRef is then used to look up the Var for the
// variable. This step is called "deref" in the code.
//
// The resolve phase can take place during evaluation as well for introspection.

// Keeps all the information statically about a variable referenced by a
// qualified name.
type varRef struct {
	scope    varScope
	index    int
	subNames []string
}

type varScope int

const (
	localScope varScope = 1 + iota
	captureScope
	builtinScope
	envScope
	externalScope
)

// An interface satisfied by both *compiler and *Frame. Used to implement
// resolveVarRef as a function that works for both types.
type scopeSearcher interface {
	searchLocal(k string) int
	searchCapture(k string) int
	searchBuiltin(k string, r diag.Ranger) int
}

// Resolves a qname into a varRef.
func resolveVarRef(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	qname = strings.TrimPrefix(qname, ":")
	if ref := resolveVarRefLocal(s, qname); ref != nil {
		return ref
	}
	if ref := resolveVarRefCapture(s, qname); ref != nil {
		return ref
	}
	if ref := resolveVarRefBuiltin(s, qname, r); ref != nil {
		return ref
	}
	return nil
}

func resolveVarRefLocal(s scopeSearcher, qname string) *varRef {
	first, rest := SplitQName(qname)
	index := s.searchLocal(first)
	if index != -1 {
		return &varRef{scope: localScope, index: index, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefCapture(s scopeSearcher, qname string) *varRef {
	first, rest := SplitQName(qname)
	if index := s.searchCapture(first); index != -1 {
		return &varRef{scope: captureScope, index: index, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefBuiltin(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	first, rest := SplitQName(qname)
	if rest != "" {
		// Try the 5 special namespace that we pretend are subnamespaces of
		// builtin:.
		switch first {
		case "local:":
			return resolveVarRefLocal(s, rest)
		case "up:":
			return resolveVarRefCapture(s, rest)
		case "builtin:":
			return resolveVarRefBuiltin(s, rest, r)
		case "e:":
			if strings.HasSuffix(rest, FnSuffix) {
				return &varRef{scope: externalScope, subNames: []string{rest[:len(rest)-1]}}
			}
		case "E:":
			return &varRef{scope: envScope, subNames: []string{rest}}
		}
	}
	if index := s.searchBuiltin(first, r); index != -1 {
		return &varRef{scope: builtinScope, index: index, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

// Tries to resolve the command head as an internal command, i.e. a builtin
// special command or a function.
func resolveCmdHeadInternally(s scopeSearcher, head string, r diag.Ranger) (compileBuiltin, *varRef) {
	special, ok := builtinSpecials[head]
	if ok {
		return special, nil
	}
	sigil, qname := SplitSigil(head)
	if sigil == "" {
		varName := qname + FnSuffix
		ref := resolveVarRef(s, varName, r)
		if ref != nil {
			return nil, ref
		}
	}
	return nil, nil
}

// Dereferences a varRef into a Var.
func deref(fm *Frame, ref *varRef) vars.Var {
	variable, subNames := derefBase(fm, ref)
	for _, subName := range subNames {
		ns, ok := variable.Get().(*Ns)
		if !ok {
			return nil
		}
		variable = ns.IndexName(subName)
		if variable == nil {
			return nil
		}
	}
	return variable
}

func derefBase(fm *Frame, ref *varRef) (vars.Var, []string) {
	switch ref.scope {
	case localScope:
		return fm.local.slots[ref.index], ref.subNames
	case captureScope:
		return fm.up.slots[ref.index], ref.subNames
	case builtinScope:
		return fm.Evaler.Builtin().slots[ref.index], ref.subNames
	case envScope:
		return vars.FromEnv(ref.subNames[0]), nil
	case externalScope:
		return vars.NewReadOnly(NewExternalCmd(ref.subNames[0])), nil
	default:
		return nil, nil
	}
}

func (cp *compiler) searchLocal(k string) int {
	return cp.thisScope().lookup(k)
}

func (cp *compiler) searchCapture(k string) int {
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		index := cp.scopes[i].lookup(k)
		if index != -1 {
			// Record the capture from i+1 to len(cp.scopes)-1, and reuse the
			// index to keep the index into the previous scope.
			index = cp.captures[i+1].add(k, true, index)
			for j := i + 2; j < len(cp.scopes); j++ {
				index = cp.captures[j].add(k, false, index)
			}
			return index
		}
	}
	return -1
}

func (cp *compiler) searchBuiltin(k string, r diag.Ranger) int {
	index := cp.builtin.lookup(k)
	if index != -1 {
		cp.checkDeprecatedBuiltin(k, r)
	}
	return index
}

func (cp *compiler) checkDeprecatedBuiltin(name string, r diag.Ranger) {
	if cp.warn == nil || r == nil {
		return
	}
	msg := ""
	switch name {
	case "-source~":
		msg = `the "source" command is deprecated; use "eval" instead`
	case "ord~":
		msg = `the "ord" command is deprecated; use "str:to-codepoints" instead`
	case "chr~":
		msg = `the "chr" command is deprecated; use "str:from-codepoints" instead`
	case "has-prefix~":
		msg = `the "has-prefix" command is deprecated; use "str:has-prefix" instead`
	case "has-suffix~":
		msg = `the "has-suffix" command is deprecated; use "str:has-suffix" instead`
	case "esleep~":
		msg = `the "esleep" command is deprecated; use "sleep" instead`
	case "eval-symlinks~":
		msg = `the "eval-symlinks" command is deprecated; use "path:eval-symlinks" instead`
	case "path-abs~":
		msg = `the "path-abs" command is deprecated; use "path:abs" instead`
	case "path-base~":
		msg = `the "path-base" command is deprecated; use "path:base" instead`
	case "path-clean~":
		msg = `the "path-clean" command is deprecated; use "path:clean" instead`
	case "path-dir~":
		msg = `the "path-dir" command is deprecated; use "path:dir" instead`
	case "path-ext~":
		msg = `the "path-ext" command is deprecated; use "path:ext" instead`
	case "-is-dir~":
		msg = `the "-is-dir" command is deprecated; use "path:is-dir" instead`
	default:
		return
	}
	dep := deprecation{cp.srcMeta.Name, r.Range(), msg}
	if prog.ShowDeprecations && cp.deprecations.register(dep) {
		err := diag.Error{
			Type: "deprecation", Message: msg,
			Context: diag.Context{
				Name: cp.srcMeta.Name, Source: cp.srcMeta.Code, Ranging: r.Range()}}
		fmt.Fprintln(cp.warn, err.Show(""))
	}
}

func (fm *Frame) searchLocal(k string) int {
	return fm.local.lookup(k)
}

func (fm *Frame) searchCapture(k string) int {
	return fm.up.lookup(k)
}

func (fm *Frame) searchBuiltin(k string, r diag.Ranger) int {
	return fm.Evaler.Builtin().lookup(k)
}
