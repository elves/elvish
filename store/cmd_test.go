package store

import "testing"

var (
	cmds     = []string{"echo foo", "put bar", "put lorem", "echo bar"}
	searches = []struct {
		first     bool
		seq       int
		prefix    string
		wantedCmd Cmd
		wantedErr error
	}{
		{false, 5, "echo", Cmd{4, "echo bar"}, nil},
		{false, 5, "put", Cmd{3, "put lorem"}, nil},
		{false, 4, "echo", Cmd{1, "echo foo"}, nil},
		{false, 3, "f", Cmd{0, ""}, ErrNoMatchingCmd},

		{true, 1, "echo", Cmd{1, "echo foo"}, nil},
		{true, 1, "put", Cmd{2, "put bar"}, nil},
		{true, 2, "echo", Cmd{4, "echo bar"}, nil},
		{true, 4, "put", Cmd{0, ""}, ErrNoMatchingCmd},
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
		f := tStore.GetLastCmd
		fname := "tStore.LastCmd"
		if tt.first {
			f = tStore.GetFirstCmd
			fname = "tStore.FirstCmd"
		}
		cmd, err := f(tt.seq, tt.prefix)
		if cmd != tt.wantedCmd || err != tt.wantedErr {
			t.Errorf("%s(%v, %v) => (%v, %v), want (%v, %v)",
				fname, tt.seq, tt.prefix,
				cmd, err, tt.wantedCmd, tt.wantedErr)
		}
	}
}
