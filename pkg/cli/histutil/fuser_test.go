package histutil

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/store"
)

var mockError = errors.New("mock error")

func TestNewFuser_Error(t *testing.T) {
	_, err := NewFuser(&TestDB{OneOffError: mockError})
	if err != mockError {
		t.Errorf("NewFuser -> error %v, want %v", err, mockError)
	}
}

func TestFusuer_AddCmd(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)
	f.AddCmd("session 1")

	wantDBCmds := []string{"store 1", "session 1"}
	wantSessionCmds := []store.Cmd{{"session 1", 1}}
	if !reflect.DeepEqual(db.AllCmds, wantDBCmds) {
		t.Errorf("DB commands = %v, want %v", db.AllCmds, wantDBCmds)
	}
	sessionCmds := f.SessionCmds()
	if !reflect.DeepEqual(sessionCmds, wantSessionCmds) {
		t.Errorf("Session commands = %v, want %v", sessionCmds, wantSessionCmds)
	}
}

func TestFuser_AddCmd_Error(t *testing.T) {
	db := &TestDB{}
	f := mustNewFuser(db)
	db.OneOffError = mockError

	_, err := f.AddCmd("haha")

	if err != mockError {
		t.Errorf("AddCmd -> error %v, want %v", err, mockError)
	}
	if len(f.SessionCmds()) != 0 {
		t.Errorf("AddCmd adds to session commands when erroring")
	}
	if len(f.SessionCmds()) != 0 {
		t.Errorf("AddCmd adds to all commands when erroring")
	}
}

func TestFuser_AllCmds(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)

	// Simulate adding commands from both the current session and other sessions.
	f.AddCmd("session 1")
	db.AddCmd("other session 1")
	db.AddCmd("other session 2")
	f.AddCmd("session 2")
	db.AddCmd("other session 3")

	// AllCmds should return all commands from the storage when the Fuser was
	// created, plus session commands. The session commands should have sequence
	// numbers consistent with the DB.
	cmds, err := f.AllCmds()
	if err != nil {
		t.Errorf("AllCmds -> error %v, want nil", err)
	}
	wantCmds := []store.Cmd{
		{"store 1", 0}, {"session 1", 1}, {"session 2", 4}}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("AllCmds -> %v, want %v", cmds, wantCmds)
	}
}

func TestFuser_AllCmds_Error(t *testing.T) {
	db := &TestDB{}
	f := mustNewFuser(db)
	db.OneOffError = mockError

	_, err := f.AllCmds()

	if err != mockError {
		t.Errorf("AllCmds -> error %v, want %v", err, mockError)
	}
}

func TestFuser_LastCmd_FromDB(t *testing.T) {
	f := mustNewFuser(&TestDB{AllCmds: []string{"store 1"}})

	cmd, _ := f.LastCmd()

	wantCmd := store.Cmd{"store 1", 0}
	if cmd != wantCmd {
		t.Errorf("LastCmd -> %v, want %v", cmd, wantCmd)
	}
}

func TestFuser_LastCmd_FromDB_Error(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)

	db.OneOffError = mockError
	_, err := f.LastCmd()

	if err != mockError {
		t.Errorf("LastCmd -> error %v, want %v", err, mockError)
	}
}

func TestFuser_LastCmd_FromSession(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)
	f.AddCmd("session 1")

	// LastCmd does not use DB when there are any session commands.
	db.OneOffError = mockError
	cmd, _ := f.LastCmd()

	wantCmd := store.Cmd{"session 1", 1}
	if cmd != wantCmd {
		t.Errorf("LastCmd -> %v, want %v", cmd, wantCmd)
	}
}

func TestFuser_FastForward(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)

	// Simulate adding commands from both the current session and other sessions.
	f.AddCmd("session 1")
	db.AddCmd("other session 1")
	db.AddCmd("other session 2")
	f.AddCmd("session 2")
	db.AddCmd("other session 3")

	f.FastForward()
	db.AddCmd("other session 4")

	wantCmds := []store.Cmd{
		{"store 1", 0}, {"session 1", 1}, {"other session 1", 2},
		{"other session 2", 3}, {"session 2", 4}, {"other session 3", 5}}
	cmds, _ := f.AllCmds()
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("AllCmds -> %v, want %v", cmds, wantCmds)
	}
}

func TestFuser_Walker(t *testing.T) {
	db := &TestDB{AllCmds: []string{"store 1"}}
	f := mustNewFuser(db)

	// Simulate adding commands from both the current session and other sessions.
	f.AddCmd("session 1")
	db.AddCmd("other session 1")
	db.AddCmd("other session 2")
	f.AddCmd("session 2")
	db.AddCmd("other session 3")

	// Walker should return a walker that walks through shared and session
	// commands.
	w := f.Walker("")
	w.Prev()
	checkWalkerCurrent(t, w, 4, "session 2")
	w.Prev()
	checkWalkerCurrent(t, w, 1, "session 1")
	w.Prev()
	checkWalkerCurrent(t, w, 0, "store 1")
	checkError(t, w.Prev(), ErrEndOfHistory)
}

func mustNewFuser(db DB) *Fuser {
	f, err := NewFuser(db)
	if err != nil {
		panic(err)
	}
	return f
}
