package testpkg

import "src.elv.sh/cmd/structoffunc/testpkg/os"

// In a separate file,
// with a different import with the name "os".
type J interface {
	A() os.File
}
