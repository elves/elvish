package listing

// State keeps the state of the listing mode.
type State struct {
	itemsGetter func(string) Items
	filtering   bool
	filter      string
	items       Items
	first       int
	selected    int
}

func (st *State) refilter(f string) {
	st.filter = f
	if st.itemsGetter == nil {
		st.items = sliceItems{}
	} else {
		st.items = st.itemsGetter(f)
	}
}

// Up moves the selection up.
func (st *State) Up() {
	if st.selected > 0 {
		st.selected--
	}
}

// UpCycle moves the selection up, wrapping to the last item if the currently
// selected item is the first item.
func (st *State) UpCycle() {
	if st.selected > 0 {
		st.selected--
	} else {
		st.selected = st.items.Len() - 1
	}
}

// Down moves the selection down.
func (st *State) Down() {
	if st.selected < st.items.Len()-1 {
		st.selected++
	}
}

// DownCycle moves the selection down, wrapping to the first item if the
// currently selected item is the last item.
func (st *State) DownCycle() {
	if st.selected < st.items.Len()-1 {
		st.selected++
	} else {
		st.selected = 0
	}
}

// ToggleFiltering toggles the filtering status of the state.
func (st *State) ToggleFiltering() {
	st.filtering = !st.filtering
}
