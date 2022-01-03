package eval

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vars"
)

// This file implements variable resolution. Elvish has fully static lexical
// scopes, so variable resolution involves some work in the compilation phase as
// well.
//
// During compilation, a qualified variable name (whether in lvalue, like "x
// = foo", or in variable use, like "$x") is searched in compiler's staticNs
// tables to determine which scope they belong to, as well as their indices in
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
	info     staticVarInfo
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
	searchLocal(k string) (staticVarInfo, int)
	searchCapture(k string) (staticVarInfo, int)
	searchBuiltin(k string, r diag.Ranger) (staticVarInfo, int)
}

// Resolves a qname into a varRef.
func resolveVarRef(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	if strings.HasPrefix(qname, ":") {
		// $:foo is reserved for fully-qualified names in future
		return nil
	}
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
	if info, index := s.searchLocal(first); index != -1 {
		return &varRef{localScope, info, index, SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefCapture(s scopeSearcher, qname string) *varRef {
	first, rest := SplitQName(qname)
	if info, index := s.searchCapture(first); index != -1 {
		return &varRef{captureScope, info, index, SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefBuiltin(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	first, rest := SplitQName(qname)
	if rest != "" {
		// Try special namespace first.
		switch first {
		case "e:":
			if strings.HasSuffix(rest, FnSuffix) {
				return &varRef{scope: externalScope, subNames: []string{rest[:len(rest)-1]}}
			}
		case "E:":
			return &varRef{scope: envScope, subNames: []string{rest}}
		}
	}
	if info, index := s.searchBuiltin(first, r); index != -1 {
		return &varRef{builtinScope, info, index, SplitQNameSegs(rest)}
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
		variable = ns.IndexString(subName)
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

func (cp *compiler) searchLocal(k string) (staticVarInfo, int) {
	return cp.thisScope().lookup(k)
}

func (cp *compiler) searchCapture(k string) (staticVarInfo, int) {
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		info, index := cp.scopes[i].lookup(k)
		if index != -1 {
			// Record the capture from i+1 to len(cp.scopes)-1, and reuse the
			// index to keep the index into the previous scope.
			index = cp.captures[i+1].add(k, true, index)
			for j := i + 2; j < len(cp.scopes); j++ {
				index = cp.captures[j].add(k, false, index)
			}
			return info, index
		}
	}
	return staticVarInfo{}, -1
}

func (cp *compiler) searchBuiltin(k string, r diag.Ranger) (staticVarInfo, int) {
	info, index := cp.builtin.lookup(k)
	if index != -1 {
		cp.checkDeprecatedBuiltin(k, r)
	}
	return info, index
}

func (fm *Frame) searchLocal(k string) (staticVarInfo, int) {
	return fm.local.lookup(k)
}

func (fm *Frame) searchCapture(k string) (staticVarInfo, int) {
	return fm.up.lookup(k)
}

func (fm *Frame) searchBuiltin(k string, r diag.Ranger) (staticVarInfo, int) {
	return fm.Evaler.Builtin().lookup(k)
}
