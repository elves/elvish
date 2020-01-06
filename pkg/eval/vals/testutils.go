package vals

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

// Want keeps wanted properties of a value.
type Want struct {
	Kind string
	Hash uint32
	Repr string

	EqualTo    interface{}
	NotEqualTo []interface{}
}

// TestValue tests properties of a value.
func TestValue(t *testing.T, v interface{}, want Want) {
	t.Helper()
	if want.Kind != "" {
		if kind := Kind(v); kind != want.Kind {
			t.Errorf("Kind() = %s, want %s", kind, want.Kind)
		}
	}
	if want.Hash != 0 {
		if hash := Hash(v); hash != want.Hash {
			t.Errorf("Hash() = %v, want %v", hash, want.Hash)
		}
	}
	if want.Repr != "" {
		if repr := Repr(v, -1); repr != want.Repr {
			t.Errorf("Repr() = %q, want %q", repr, want.Repr)
		}
	}
	if want.EqualTo != nil {
		if eq := Equal(v, want.EqualTo); !eq {
			t.Errorf("Equal(%v) = false, want true", want.EqualTo)
		}
	}
	for _, other := range want.NotEqualTo {
		if eq := Equal(v, other); eq {
			t.Errorf("Equal(%v) = true, want false", other)
		}
	}
}

// Eq returns a tt.Matcher that matches using the Equal function.
func Eq(r interface{}) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want interface{} }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }
