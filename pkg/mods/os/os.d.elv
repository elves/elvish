#doc:show-unstable
#
# Reports whether an exception is caused by the fact that a file or directory
# already exists.
fn -is-exist {|exc| }

#doc:show-unstable
#
# Reports whether an exception is caused by the fact that a file or directory
# does not exist.
fn -is-not-exist {|exc| }

# Creates a new directory with the specified name and permission (before umask).
fn mkdir {|&perm=0o755 path| }

# Removes the file or empty directory at `path`.
#
# If the path does not exist, this command throws an exception that can be
# tested with [`os:-is-not-exist`]().
fn remove {|path| }

# Removes the named file or directory at `path` and, in the latter case, any
# children it contains. It removes everything it can, but returns the first
# error it encounters.
#
# If the path does not exist, this command returns silently without throwing an
# exception.
fn remove-all {|path| }

# Reports whether a file is known to exist at `path`.
fn exists {|path| }
