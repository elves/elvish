package listbox

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Itemer    Itemer
	NItems    int
	Selected  int
	LastFirst int
}
