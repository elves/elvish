package store

import "testing"

var (
	lines    = []string{"echo foo", "put bar", "put lorem", "echo bar"}
	searches = []struct {
		maxSeq     int
		prefix     string
		wantedSeq  int
		wantedLine string
		wantedErr  error
	}{
		{4, "echo", 4, "echo bar", nil},
		{4, "put", 3, "put lorem", nil},
		{3, "echo", 1, "echo foo", nil},
		{4, "f", 0, "", ErrNoMatchingLine},
	}
)

func TestLine(t *testing.T) {
	startSeq, err := tStore.GetMaxLineSeq()
	if startSeq != 0 || err != nil {
		t.Errorf("tStore.GetLastLineSeq() => (%v, %v), want (0, nil)",
			startSeq, err)
	}
	for _, line := range lines {
		err := tStore.AddLine(line)
		if err != nil {
			t.Errorf("tStore.AddLine(%v) => %v, want nil", line, err)
		}
	}
	endSeq, err := tStore.GetMaxLineSeq()
	wantedEndSeq := startSeq + len(lines)
	if endSeq != wantedEndSeq || err != nil {
		t.Errorf("tStore.GetLastLineSeq() => (%v, %v), want (%v, nil)",
			endSeq, err, wantedEndSeq)
	}
	for i, wantedLine := range lines {
		seq := i + startSeq + 1
		line, err := tStore.GetLine(seq)
		if line != wantedLine || err != nil {
			t.Errorf("tStore.GetLine(%v) => (%v, %v), want (%v, nil)",
				seq, line, err, wantedLine)
		}
	}
	for _, f := range searches {
		seq, line, err := tStore.GetLastLineWithPrefix(f.maxSeq, f.prefix)
		if seq != f.wantedSeq || line != f.wantedLine || err != f.wantedErr {
			t.Errorf("tStore.GetLastLineWithPrefix(%v, %v) => (%v, %v, %v), want (%v, %v, %v)",
				f.maxSeq, f.prefix, seq, line, err, f.wantedSeq, f.wantedLine, f.wantedErr)
		}
	}
}
