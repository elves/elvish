package vals

// A test structmap type used in other tests.
type testStructMap struct {
	Name        string  `json:"name"`
	ScoreNumber float64 `json:"score-number"`
}

func (testStructMap) IsStructMap(StructMapMarker) {}
