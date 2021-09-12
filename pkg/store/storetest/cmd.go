package storetest

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/store/storedefs"
)

var (
	cmds     = []string{"echo foo", "put bar", "put lorem", "echo bar"}
	searches = []struct {
		next      bool
		seq       int
		prefix    string
		wantedSeq int
		wantedCmd string
		wantedErr error
	}{
		{false, 5, "echo", 4, "echo bar", nil},
		{false, 5, "put", 3, "put lorem", nil},
		{false, 4, "echo", 1, "echo foo", nil},
		{false, 3, "f", 0, "", storedefs.ErrNoMatchingCmd},
		{false, 1, "", 0, "", storedefs.ErrNoMatchingCmd},

		{true, 1, "echo", 1, "echo foo", nil},
		{true, 1, "put", 2, "put bar", nil},
		{true, 2, "echo", 4, "echo bar", nil},
		{true, 4, "put", 0, "", storedefs.ErrNoMatchingCmd},
	}
)

// TestCmd tests the command history functionality of a Store.
func TestCmd(t *testing.T, store storedefs.Store) {
	startSeq, err := store.NextCmdSeq()
	if startSeq != 1 || err != nil {
		t.Errorf("store.NextCmdSeq() => (%v, %v), want (1, nil)",
			startSeq, err)
	}

	// AddCmd
	for i, cmd := range cmds {
		wantSeq := startSeq + i
		seq, err := store.AddCmd(cmd)
		if seq != wantSeq || err != nil {
			t.Errorf("store.AddCmd(%v) => (%v, %v), want (%v, nil)",
				cmd, seq, err, wantSeq)
		}
	}

	endSeq, err := store.NextCmdSeq()
	wantedEndSeq := startSeq + len(cmds)
	if endSeq != wantedEndSeq || err != nil {
		t.Errorf("store.NextCmdSeq() => (%v, %v), want (%v, nil)",
			endSeq, err, wantedEndSeq)
	}

	// CmdsWithSeq
	wantCmdWithSeqs := make([]storedefs.Cmd, len(cmds))
	for i, cmd := range cmds {
		wantCmdWithSeqs[i] = storedefs.Cmd{Text: cmd, Seq: i + 1}
	}
	for i := 0; i < len(cmds); i++ {
		for j := i; j <= len(cmds); j++ {
			cmdWithSeqs, err := store.CmdsWithSeq(i+1, j+1)
			if !equalCmds(cmdWithSeqs, wantCmdWithSeqs[i:j]) || err != nil {
				t.Errorf("store.CmdsWithSeq(%v, %v) -> (%v, %v), want (%v, nil)",
					i+1, j+1, cmdWithSeqs, err, wantCmdWithSeqs[i:j])
			}
		}
	}

	// Cmd
	for i, wantedCmd := range cmds {
		seq := i + startSeq
		cmd, err := store.Cmd(seq)
		if cmd != wantedCmd || err != nil {
			t.Errorf("store.Cmd(%v) => (%v, %v), want (%v, nil)",
				seq, cmd, err, wantedCmd)
		}
	}

	// PrevCmd and NextCmd
	for _, tt := range searches {
		f := store.PrevCmd
		funcname := "store.PrevCmd"
		if tt.next {
			f = store.NextCmd
			funcname = "store.NextCmd"
		}
		cmd, err := f(tt.seq, tt.prefix)
		wantedCmd := storedefs.Cmd{Text: tt.wantedCmd, Seq: tt.wantedSeq}
		if cmd != wantedCmd || !matchErr(err, tt.wantedErr) {
			t.Errorf("%s(%v, %v) => (%v, %v), want (%v, %v)",
				funcname, tt.seq, tt.prefix, cmd, err, wantedCmd, tt.wantedErr)
		}
	}

	// DelCmd
	if err := store.DelCmd(1); err != nil {
		t.Error("Failed to remove cmd")
	}
	if seq, err := store.Cmd(1); !matchErr(err, storedefs.ErrNoMatchingCmd) {
		t.Errorf("Cmd(1) => (%v, %v), want (%v, %v)",
			seq, err, "", storedefs.ErrNoMatchingCmd)
	}
}

func equalCmds(a, b []storedefs.Cmd) bool {
	return (len(a) == 0 && len(b) == 0) || reflect.DeepEqual(a, b)
}
