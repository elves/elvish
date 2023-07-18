module src.elv.sh/website

go 1.19

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/creack/pty v1.1.18
	github.com/google/go-cmp v0.5.9
	src.elv.sh v0.19.2
)

require (
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)

replace src.elv.sh => ../
