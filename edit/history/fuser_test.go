package history

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewFuser(t *testing.T) {
	mockError := errors.New("mock error")
	_, err := NewFuser(&mockStore{oneOffError: mockError})
	if err != mockError {
		t.Errorf("NewFuser -> error %v, want %v", err, mockError)
	}
}

var fuserStore = &mockStore{cmds: []string{"ls -x"}}

func TestFuser(t *testing.T) {
	f, err := NewFuser(fuserStore)
	if err != nil {
		t.Errorf("NewFuser -> error %v, want nil", err)
	}

	// AddCmd should not add command to session history if backend has an error
	// adding the command.
	mockError := errors.New("mock error")
	fuserStore.oneOffError = mockError
	err = f.AddCmd("haha")
	if err != mockError {
		t.Errorf("AddCmd doesn't forward backend error")
	}
	if len(f.cmds) != 0 {
		t.Errorf("AddCmd adds command to session history when backend errors")
	}

	// AddCmd should add command to both storage and session
	f.AddCmd("echo -n")
	if !reflect.DeepEqual(fuserStore.cmds, []string{"ls -x", "echo -n"}) {
		t.Errorf("AddCmd doesn't add command to backend storage")
	}
	if !reflect.DeepEqual(f.cmds, []string{"echo -n"}) {
		t.Errorf("AddCmd doesn't add command to session history")
	}

	// SessionCmds should return commands added in this session
	cmds := f.SessionCmds()
	if !reflect.DeepEqual(cmds, []string{"echo -n"}) {
		t.Errorf("SessionCmds doesn't return session commands")
	}

	// AllCmds should return all commands from the storage when the Fuser was
	// created followed by session commands
	fuserStore.AddCmd("blah blah")
	fuserStore.AddCmd("lorem ipsum")
	cmds, err = f.AllCmds()
	if err != nil {
		t.Errorf("AllCmds returns error")
	}
	if !reflect.DeepEqual(cmds, []string{"ls -x", "echo -n"}) {
		t.Errorf("AllCmds doesn't return all commands")
	}

	// AllCmds should forward backend storage error
	mockError = errors.New("another mock error")
	fuserStore.oneOffError = mockError
	_, err = f.AllCmds()
	if err != mockError {
		t.Errorf("AllCmds doesn't forward backend error")
	}

	// Walker should return a walker that walks through all commands
	w := f.Walker("")
	wantCmd(t, w.Prev, 9999, "echo -n")
	wantCmd(t, w.Prev, 0, "ls -x")
	wantErr(t, w.Prev, ErrEndOfHistory)
}
