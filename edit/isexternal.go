package edit

func (ed *Editor) updateIsExternal() {
	names := make(chan string, 32)
	go func() {
		ed.evaler.AllExecutables(names)
		close(names)
	}()
	isExternal := make(map[string]bool)
	for name := range names {
		isExternal[name] = true
	}
	ed.isExternal.Lock()
	ed.isExternal.m = isExternal
	ed.isExternal.Unlock()
}
