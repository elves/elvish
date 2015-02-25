package store

import "testing"

var (
	lines    = []string{"echo foo", "put bar", "put lorem", "echo bar"}
	searches = []struct {
		first      bool
		seq        int
		prefix     string
		wantedSeq  int
		wantedLine string
		wantedErr  error
	}{
		{false, 5, "echo", 4, "echo bar", nil},
		{false, 5, "put", 3, "put lorem", nil},
		{false, 4, "echo", 1, "echo foo", nil},
		{false, 3, "f", 0, "", ErrNoMatchingLine},

		{true, 1, "echo", 1, "echo foo", nil},
		{true, 1, "put", 2, "put bar", nil},
		{true, 2, "echo", 4, "echo bar", nil},
		{true, 4, "put", 0, "", ErrNoMatchingLine},
	}
)

func TestLine(t *testing.T) {
	startSeq, err := tStore.GetNextLineSeq()
	if startSeq != 1 || err != nil {
		t.Errorf("tStore.GetNextLineSeq() => (%v, %v), want (1, nil)",
			startSeq, err)
	}
	for _, line := range lines {
		err := tStore.AddLine(line)
		if err != nil {
			t.Errorf("tStore.AddLine(%v) => %v, want nil", line, err)
		}
	}
	endSeq, err := tStore.GetNextLineSeq()
	wantedEndSeq := startSeq + len(lines)
	if endSeq != wantedEndSeq || err != nil {
		t.Errorf("tStore.GetLastLineSeq() => (%v, %v), want (%v, nil)",
			endSeq, err, wantedEndSeq)
	}
	for i, wantedLine := range lines {
		seq := i + startSeq
		line, err := tStore.GetLine(seq)
		if line != wantedLine || err != nil {
			t.Errorf("tStore.GetLine(%v) => (%v, %v), want (%v, nil)",
				seq, line, err, wantedLine)
		}
	}
	for _, tt := range searches {
		f := tStore.GetLastLineWithPrefix
		fname := "tStore.GetLastLineWithPrefix"
		if tt.first {
			f = tStore.GetFirstLineWithPrefix
			fname = "tStore.GetFirstLineWithPrefix"
		}
		seq, line, err := f(tt.seq, tt.prefix)
		if seq != tt.wantedSeq || line != tt.wantedLine || err != tt.wantedErr {
			t.Errorf("%s(%v, %v) => (%v, %v, %v), want (%v, %v, %v)",
				fname, tt.seq, tt.prefix,
				seq, line, err,
				tt.wantedSeq, tt.wantedLine, tt.wantedErr)
		}
	}
}
