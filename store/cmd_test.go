package store

import "testing"

var (
	cmds     = []string{"echo foo", "put bar", "put lorem", "echo bar"}
	searches = []struct {
		first     bool
		seq       int
		prefix    string
		wantedSeq int
		wantedCmd string
		wantedErr error
	}{
		{false, 5, "echo", 4, "echo bar", nil},
		{false, 5, "put", 3, "put lorem", nil},
		{false, 4, "echo", 1, "echo foo", nil},
		{false, 3, "f", 0, "", ErrNoMatchingCmd},

		{true, 1, "echo", 1, "echo foo", nil},
		{true, 1, "put", 2, "put bar", nil},
		{true, 2, "echo", 4, "echo bar", nil},
		{true, 4, "put", 0, "", ErrNoMatchingCmd},
	}
)

func TestCmd(t *testing.T) {
	startSeq, err := tStore.NextCmdSeq()
	if startSeq != 1 || err != nil {
		t.Errorf("tStore.NextCmdSeq() => (%v, %v), want (1, nil)",
			startSeq, err)
	}
	for _, cmd := range cmds {
		err := tStore.AddCmd(cmd)
		if err != nil {
			t.Errorf("tStore.AddCmd(%v) => %v, want nil", cmd, err)
		}
	}
	endSeq, err := tStore.NextCmdSeq()
	wantedEndSeq := startSeq + len(cmds)
	if endSeq != wantedEndSeq || err != nil {
		t.Errorf("tStore.NextCmdSeq() => (%v, %v), want (%v, nil)",
			endSeq, err, wantedEndSeq)
	}
	for i, wantedCmd := range cmds {
		seq := i + startSeq
		cmd, err := tStore.GetCmd(seq)
		if cmd != wantedCmd || err != nil {
			t.Errorf("tStore.Cmd(%v) => (%v, %v), want (%v, nil)",
				seq, cmd, err, wantedCmd)
		}
	}
	for _, tt := range searches {
		f := tStore.LastCmd
		fname := "tStore.LastCmd"
		if tt.first {
			f = tStore.FirstCmd
			fname = "tStore.FirstCmd"
		}
		seq, cmd, err := f(tt.seq, tt.prefix)
		if seq != tt.wantedSeq || cmd != tt.wantedCmd || err != tt.wantedErr {
			t.Errorf("%s(%v, %v) => (%v, %v, %v), want (%v, %v, %v)",
				fname, tt.seq, tt.prefix,
				seq, cmd, err,
				tt.wantedSeq, tt.wantedCmd, tt.wantedErr)
		}
	}
}
