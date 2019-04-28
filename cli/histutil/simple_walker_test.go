package histutil

import "testing"

type step struct {
	f       func(Walker) error
	wantSeq int
	wantCmd string
	wantErr error
}

var (
	prev = Walker.Prev
	next = Walker.Next
)

var simpleWalkerTests = []struct {
	prefix string
	cmds   []string
	steps  []step
}{
	{
		"",
		[]string{},
		[]step{
			{next, 0, "", ErrEndOfHistory},
			{prev, 0, "", ErrEndOfHistory},
		},
	},
	{
		"",
		//       0        1        2         3        4         5
		[]string{"ls -l", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
		[]step{
			{prev, 5, "ls a", nil},
			{next, 6, "", ErrEndOfHistory}, // Next does not stop at border
			{prev, 5, "ls a", nil},
			{prev, 4, "echo a", nil},
			{prev, 3, "ls -a", nil},
			{prev, 1, "ls -l", nil},             // skips 2; dup with 4
			{prev, 1, "ls -l", ErrEndOfHistory}, // Prev stops at border
			{next, 3, "ls -a", nil},
			{next, 4, "echo a", nil},
			{next, 5, "ls a", nil},
		},
	},
	{
		"e",
		//       0         1        2         3        4         5
		[]string{"echo a", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
		[]step{
			{prev, 4, "echo a", nil},
			{prev, 4, "echo a", ErrEndOfHistory},
			{next, 6, "", ErrEndOfHistory},
		},
	},
	{
		"l",
		//       0         1        2         3        4         5
		[]string{"echo a", "ls -l", "echo a", "ls -a", "echo a", "ls a"},
		[]step{
			{prev, 5, "ls a", nil},
			{prev, 3, "ls -a", nil},
			{prev, 1, "ls -l", nil},
			{prev, 1, "ls -l", ErrEndOfHistory},
			{next, 3, "ls -a", nil},
			{next, 5, "ls a", nil},
			{next, 6, "", ErrEndOfHistory},
		},
	},
}

func TestSimpleWalker(t *testing.T) {
	for _, test := range simpleWalkerTests {
		w := NewSimpleWalker(test.cmds, test.prefix)
		for _, step := range test.steps {
			err := step.f(w)
			if err != step.wantErr {
				t.Errorf("Got error %v, want %v", err, step.wantErr)
			}
			seq := w.CurrentSeq()
			if seq != step.wantSeq {
				t.Errorf("Got seq %v, want %v", seq, step.wantSeq)
			}
			cmd := w.CurrentCmd()
			if cmd != step.wantCmd {
				t.Errorf("Got cmd %v, want %v", cmd, step.wantCmd)
			}
		}
	}
}
