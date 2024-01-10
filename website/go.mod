module src.elv.sh/website

go 1.20

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/creack/pty v1.1.21
	github.com/google/go-cmp v0.6.0
	src.elv.sh v0.19.2
)

require golang.org/x/sys v0.16.0 // indirect

replace src.elv.sh => ../
