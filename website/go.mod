module src.elv.sh/website

go 1.19

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/creack/pty v1.1.18
	github.com/google/go-cmp v0.5.9
	src.elv.sh v0.18.0
)

require (
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

replace src.elv.sh => ../
