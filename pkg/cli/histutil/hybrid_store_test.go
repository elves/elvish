package histutil

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/store"
)

var mockError = errors.New("mock error")

func TestNewHybridStore_ReturnsMemStoreIfDBIsNil(t *testing.T) {
	store, err := NewHybridStore(nil)
	if _, ok := store.(*memStore); !ok {
		t.Errorf("NewHybridStore -> %v, want memStore", store)
	}
	if err != nil {
		t.Errorf("NewHybridStore -> error %v, want nil", err)
	}
}

func TestNewHybridStore_ReturnsMemStoreOnDBError(t *testing.T) {
	db := NewFaultyInMemoryDB()
	db.SetOneOffError(mockError)
	store, err := NewHybridStore(db)
	if _, ok := store.(*memStore); !ok {
		t.Errorf("NewHybridStore -> %v, want memStore", store)
	}
	if err != mockError {
		t.Errorf("NewHybridStore -> error %v, want %v", err, mockError)
	}
}

func TestFusuer_AddCmd_AddsBothToDBAndSession(t *testing.T) {
	db := NewFaultyInMemoryDB("shared 1")
	f := mustNewHybridStore(db)

	f.AddCmd(store.Cmd{Text: "session 1"})

	wantDBCmds := []store.Cmd{
		{Text: "shared 1", Seq: 0}, {Text: "session 1", Seq: 1}}
	if dbCmds, _ := db.CmdsWithSeq(-1, -1); !reflect.DeepEqual(dbCmds, wantDBCmds) {
		t.Errorf("DB commands = %v, want %v", dbCmds, wantDBCmds)
	}

	allCmds, err := f.AllCmds()
	if err != nil {
		panic(err)
	}
	wantAllCmds := []store.Cmd{
		{Text: "shared 1", Seq: 0},
		{Text: "session 1", Seq: 1}}
	if !reflect.DeepEqual(allCmds, wantAllCmds) {
		t.Errorf("AllCmd -> %v, want %v", allCmds, wantAllCmds)
	}
}

func TestHybridStore_AddCmd_AddsToSessionEvenIfDBErrors(t *testing.T) {
	db := NewFaultyInMemoryDB()
	f := mustNewHybridStore(db)
	db.SetOneOffError(mockError)

	_, err := f.AddCmd(store.Cmd{Text: "haha"})
	if err != mockError {
		t.Errorf("AddCmd -> error %v, want %v", err, mockError)
	}

	allCmds, err := f.AllCmds()
	if err != nil {
		panic(err)
	}
	wantAllCmds := []store.Cmd{{Text: "haha", Seq: 1}}
	if !reflect.DeepEqual(allCmds, wantAllCmds) {
		t.Errorf("AllCmd -> %v, want %v", allCmds, wantAllCmds)
	}
}

func TestHybridStore_AllCmds_IncludesFrozenSharedAndNewlyAdded(t *testing.T) {
	db := NewFaultyInMemoryDB("shared 1")
	f := mustNewHybridStore(db)

	// Simulate adding commands from both the current session and other sessions.
	f.AddCmd(store.Cmd{Text: "session 1"})
	db.AddCmd("other session 1")
	db.AddCmd("other session 2")
	f.AddCmd(store.Cmd{Text: "session 2"})
	db.AddCmd("other session 3")

	// AllCmds should return all commands from the storage when the HybridStore
	// was created, plus session commands. The session commands should have
	// sequence numbers consistent with the DB.
	allCmds, err := f.AllCmds()
	if err != nil {
		t.Errorf("AllCmds -> error %v, want nil", err)
	}
	wantAllCmds := []store.Cmd{
		{Text: "shared 1", Seq: 0},
		{Text: "session 1", Seq: 1},
		{Text: "session 2", Seq: 4}}
	if !reflect.DeepEqual(allCmds, wantAllCmds) {
		t.Errorf("AllCmds -> %v, want %v", allCmds, wantAllCmds)
	}
}

func TestHybridStore_AllCmds_ReturnsSessionIfDBErrors(t *testing.T) {
	db := NewFaultyInMemoryDB("shared 1")
	f := mustNewHybridStore(db)
	f.AddCmd(store.Cmd{Text: "session 1"})
	db.SetOneOffError(mockError)

	allCmds, err := f.AllCmds()
	if err != mockError {
		t.Errorf("AllCmds -> error %v, want %v", err, mockError)
	}
	wantAllCmds := []store.Cmd{{Text: "session 1", Seq: 1}}
	if !reflect.DeepEqual(allCmds, wantAllCmds) {
		t.Errorf("AllCmd -> %v, want %v", allCmds, wantAllCmds)
	}
}

func TestHybridStore_Cursor_OnlySession(t *testing.T) {
	db := NewFaultyInMemoryDB()
	f := mustNewHybridStore(db)
	db.AddCmd("+ other session")
	f.AddCmd(store.Cmd{Text: "+ session 1"})
	f.AddCmd(store.Cmd{Text: "- no match"})

	testCursorIteration(t, f.Cursor("+"), []store.Cmd{{Text: "+ session 1", Seq: 1}})
}

func TestHybridStore_Cursor_OnlyShared(t *testing.T) {
	db := NewFaultyInMemoryDB("- no match", "+ shared 1")
	f := mustNewHybridStore(db)
	db.AddCmd("+ other session")
	f.AddCmd(store.Cmd{Text: "- no match"})

	testCursorIteration(t, f.Cursor("+"), []store.Cmd{{Text: "+ shared 1", Seq: 1}})
}

func TestHybridStore_Cursor_SharedAndSession(t *testing.T) {
	db := NewFaultyInMemoryDB("- no match", "+ shared 1")
	f := mustNewHybridStore(db)
	db.AddCmd("+ other session")
	db.AddCmd("- no match")
	f.AddCmd(store.Cmd{Text: "+ session 1"})
	f.AddCmd(store.Cmd{Text: "- no match"})

	testCursorIteration(t, f.Cursor("+"), []store.Cmd{
		{Text: "+ shared 1", Seq: 1},
		{Text: "+ session 1", Seq: 4}})
}

func testCursorIteration(t *testing.T, cursor Cursor, wantCmds []store.Cmd) {
	expectEndOfHistory := func() {
		t.Helper()
		if _, err := cursor.Get(); err != ErrEndOfHistory {
			t.Errorf("Get -> error %v, want ErrEndOfHistory", err)
		}
	}
	expectCmd := func(i int) {
		t.Helper()
		wantCmd := wantCmds[i]
		cmd, err := cursor.Get()
		if cmd != wantCmd {
			t.Errorf("Get -> %v, want %v", cmd, wantCmd)
		}
		if err != nil {
			t.Errorf("Get -> error %v, want nil", err)
		}
	}

	expectEndOfHistory()

	for i := len(wantCmds) - 1; i >= 0; i-- {
		cursor.Prev()
		expectCmd(i)
	}

	cursor.Prev()
	expectEndOfHistory()
	cursor.Prev()
	expectEndOfHistory()

	for i := range wantCmds {
		cursor.Next()
		expectCmd(i)
	}

	cursor.Next()
	expectEndOfHistory()
	cursor.Next()
	expectEndOfHistory()
}

func mustNewHybridStore(db DB) Store {
	f, err := NewHybridStore(db)
	if err != nil {
		panic(err)
	}
	return f
}
