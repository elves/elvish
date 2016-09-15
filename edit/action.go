package edit

type action struct {
	typ        actionType
	returnLine string
	returnErr  error
}

type actionType int

const (
	noAction actionType = iota
	reprocessKey
	exitReadLine
)
