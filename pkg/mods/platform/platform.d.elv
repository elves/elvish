# The architecture of the platform; e.g. amd64, arm, ppc.
# This corresponds to Go's
# [`GOARCH`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
# This is read-only.
var arch

# The name of the operating system; e.g. darwin (macOS), linux, etc.
# This corresponds to Go's
# [`GOOS`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
# This is read-only.
var os

# Whether or not the platform is Unix-like. This includes Linux, macOS
# (Darwin), FreeBSD, NetBSD, and OpenBSD. This can be used to decide, for
# example, if the `unix` module is usable.
# This is read-only.
var is-unix

# Whether or not the platform is Microsoft Windows.
# This is read-only.
var is-windows

# Outputs the hostname of the system. If the option `&strip-domain` is `$true`,
# strips the part after the first dot.
#
# This function throws an exception if it cannot determine the hostname. It is
# implemented using Go's [`os.Hostname`](https://golang.org/pkg/os/#Hostname).
#
# Examples:
#
# ```elvish-transcript
# ~> platform:hostname
# ▶ lothlorien.elv.sh
# ~> platform:hostname &strip-domain=$true
# ▶ lothlorien
# ```
fn hostname {|&strip-domain=$false| }
