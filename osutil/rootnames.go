package osutil

import "os"

// RootNames returns the result of /*.
func RootNames() []string {
	f, err := os.Open("/")
	if err != nil {
		panic(err)
	}

	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		panic(err)
	}
	for i, name := range names {
		names[i] = "/" + name
	}
	return names
}
