This is `das`, an experimental Unix shell.

`das` consists of two parts - pipeline building, forking, etc. are written in
C; the rest in [golang](http://golang.org/). The two parts communicate using
two pipes.

`das` uses [tup](http://gittup.org/tup/) as the build system.
