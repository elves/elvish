package vals

import "testing"

type testPseudoMap struct{}

func (testPseudoMap) Kind() string      { return "test-pseudo-map" }
func (testPseudoMap) Fields() MethodMap { return methodMap{} }

type methodMap struct{}

func (methodMap) Foo() string { return "lorem" }
func (methodMap) Bar() string { return "ipsum" }
func (methodMap) FooBar() int { return 23 }

func TestPseudoMap(t *testing.T) {
	TestValue(t, testPseudoMap{}).
		Repr("[^test-pseudo-map &bar=ipsum &foo=lorem &foo-bar=(num 23)]").
		HasKey("foo", "bar", "foo-bar").
		NotEqual(
			// Pseudo maps are nominally typed, so they are not equal to maps or
			// field maps with the same pairs.
			MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23),
			fieldMap{"lorem", "ipsum", 23},
		).
		HasNoKey("bad", 1.0).
		IndexError("bad", NoSuchKey("bad")).
		IndexError(1.0, NoSuchKey(1.0)).
		AllKeys("bar", "foo", "foo-bar").
		Index("foo", "lorem").
		Index("bar", "ipsum").
		Index("foo-bar", 23)
}
