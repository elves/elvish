package vals

import (
	"testing"

	"github.com/xiaq/persistent/hash"
)

// A test structmap type used in other tests.
type testStructMap struct {
	Name        string  `json:"name"`
	ScoreNumber float64 `json:"score-number"`
}

func (testStructMap) IsStructMap(StructMapMarker) {}

func TestStructMap(t *testing.T) {
	TestValue(t, testStructMap{}).
		Kind("structmap").
		Hash(hash.DJB(Hash(""), Hash(0.0))).
		Repr(`[&name='' &score-number=(float64 0)]`).
		Len(2).
		Equal(testStructMap{}).
		NotEqual("a", MakeMap(), testStructMap{"a", 1.0}).
		HasKey("name", "score-number").
		HasNoKey("bad").
		AllKeys("name", "score-number").
		Index("name", "").
		Index("score-number", 0.0)

	TestValue(t, testStructMap{"a", 1.0}).
		Kind("structmap").
		Hash(hash.DJB(Hash("a"), Hash(1.0))).
		Repr(`[&name=a &score-number=(float64 1)]`).
		Len(2).
		Equal(testStructMap{"a", 1.0}).
		NotEqual(
			"a", MakeMap("name", "", "score-number", 1.0),
			testStructMap{}, testStructMap{"a", 2.0}, testStructMap{"b", 1.0}).
		Index("name", "a").
		Index("score-number", 1.0).
		Assoc("name", "b", testStructMap{"b", 1.0}).
		Assoc("score-number", 2.0, testStructMap{"a", 2.0}).
		AssocError("score-number", "not-num", cannotParseAs{"number", "not-num"})
}
