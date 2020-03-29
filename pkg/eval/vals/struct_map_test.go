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

type testStructMap2 struct {
	Name        string  `json:"name"`
	ScoreNumber float64 `json:"score-number"`
}

func (testStructMap2) IsStructMap(StructMapMarker) {}

func TestStructMap(t *testing.T) {
	TestValue(t, testStructMap{}).
		Kind("structmap").
		Bool(true).
		Hash(hash.DJB(Hash(""), Hash(0.0))).
		Repr(`[&name='' &score-number=(float64 0)]`).
		Len(2).
		Equal(testStructMap{}).
		NotEqual("a", MakeMap(), testStructMap{"a", 1.0}).
		// StructMap's are nominally typed. This may change in future.
		NotEqual(testStructMap2{}).
		HasKey("name", "score-number").
		HasNoKey("bad", 1.0).
		IndexError("bad", NoSuchKey("bad")).
		IndexError(1.0, NoSuchKey(1.0)).
		AllKeys("name", "score-number").
		Index("name", "").
		Index("score-number", 0.0)

	TestValue(t, testStructMap{"a", 1.0}).
		Kind("structmap").
		Bool(true).
		Hash(hash.DJB(Hash("a"), Hash(1.0))).
		Repr(`[&name=a &score-number=(float64 1)]`).
		Len(2).
		Equal(testStructMap{"a", 1.0}).
		NotEqual(
			"a", MakeMap("name", "", "score-number", 1.0),
			testStructMap{}, testStructMap{"a", 2.0}, testStructMap{"b", 1.0}).
		// Keys are tested above, thus omitted here.
		Index("name", "a").
		Index("score-number", 1.0).
		Assoc("name", "b", testStructMap{"b", 1.0}).
		Assoc("score-number", 2.0, testStructMap{"a", 2.0}).
		AssocError("score-number", "not-num", cannotParseAs{"number", "not-num"}).
		AssocError("new-key", "", errStructMapKey).
		AssocError(1.0 /* non-string key */, "", errStructMapKey)
}
