module src.elv.sh/website

go 1.22

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/creack/pty v1.1.23
	src.elv.sh v0.21.0
)

require (
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
)

replace src.elv.sh => ../
