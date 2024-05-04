module src.elv.sh/website

go 1.21

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/creack/pty v1.1.21
	src.elv.sh v0.19.2
)

require (
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace src.elv.sh => ../
