module src.elv.sh/website

go 1.18

require (
	github.com/BurntSushi/toml v1.0.0
	src.elv.sh v0.17.0
)

require golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect

replace src.elv.sh => ../
