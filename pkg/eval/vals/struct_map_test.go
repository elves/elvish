package vals

import (
	"testing"

	"src.elv.sh/pkg/persistent/hash"
)

type testStructMap struct {
	Name  string
	Score float64
}

func (testStructMap) IsStructMap() {}

func (m testStructMap) ScorePlusTen() float64 { return m.Score + 10 }

// Equivalent to testStructMap for Elvish.
type testStructMap2 struct {
	Name         string
	Score        float64
	ScorePlusTen float64
}

func (testStructMap2) IsStructMap() {}

func TestStructMap(t *testing.T) {
	TestValue(t, testStructMap{"ls", 1.0}).
		Kind("map").
		Bool(true).
		Hash(
			hash.DJB(Hash("name"), Hash("ls"))+
				hash.DJB(Hash("score"), Hash(1.0))+
				hash.DJB(Hash("score-plus-ten"), Hash(11.0))).
		Repr(`[&name=ls &score=(num 1.0) &score-plus-ten=(num 11.0)]`).
		Len(3).
		Equal(
			// Struct maps behave like maps, so they are equal to normal maps
			// and other struct maps with the same entries.
			MakeMap("name", "ls", "score", 1.0, "score-plus-ten", 11.0),
			testStructMap{"ls", 1.0},
			testStructMap2{"ls", 1.0, 11.0}).
		NotEqual("a", MakeMap(), testStructMap{"ls", 2.0}, testStructMap{"l", 1.0}).
		HasKey("name", "score", "score-plus-ten").
		HasNoKey("bad", 1.0).
		IndexError("bad", NoSuchKey("bad")).
		IndexError(1.0, NoSuchKey(1.0)).
		AllKeys("name", "score", "score-plus-ten").
		Index("name", "ls").
		Index("score", 1.0).
		Index("score-plus-ten", 11.0)
}

type testPseudoMap struct{}

func (testPseudoMap) Kind() string      { return "test-pseudo-map" }
func (testPseudoMap) Fields() StructMap { return testStructMap{"pseudo", 100} }

func TestPseudoMap(t *testing.T) {
	TestValue(t, testPseudoMap{}).
		Repr("[^test-pseudo-map &name=pseudo &score=(num 100.0) &score-plus-ten=(num 110.0)]").
		HasKey("name", "score", "score-plus-ten").
		NotEqual(
			// Pseudo struct maps are nominally typed, so they are not equal to
			// maps or struct maps with the same entries.
			MakeMap("name", "", "score", 1.0, "score-plus-ten", 11.0),
			testStructMap{"ls", 1.0},
		).
		HasNoKey("bad", 1.0).
		IndexError("bad", NoSuchKey("bad")).
		IndexError(1.0, NoSuchKey(1.0)).
		AllKeys("name", "score", "score-plus-ten").
		Index("name", "pseudo").
		Index("score", 100.0).
		Index("score-plus-ten", 110.0)
}
