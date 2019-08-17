package listbox

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Itemer    Itemer
	NItems    int
	Selected  int
	LastFirst int
}

// MakeState makes a new State.
func MakeState(it Itemer, n int, selectLast bool) State {
	selected := 0
	if selectLast {
		selected = n - 1
	}
	return State{it, n, selected, 0}
}
