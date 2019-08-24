package cliedit

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.
