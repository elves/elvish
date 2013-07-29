das, an experimental Unix shell
===============================

`das` consists of two parts.  Things related to `fork`ing are written in C
(the server); the rest in [golang](http://golang.org/).  The two parts
communicate using UNIX sockets.

`das` uses [tup](http://gittup.org/tup/) as the build system.

License
-------

BSD 2-clause license.  See LICENSE for a copy.
