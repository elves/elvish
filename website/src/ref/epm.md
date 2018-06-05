<!-- toc -->

# Introduction

The Elvish Package Manager (`epm`) is a module bundled with Elvish for
managing third-party packages.

In Elvish terminology, a **module** is a `.elv` file that can be imported with
the `use` command, while a **package** is a collection of modules that are
usually kept in the same repository as one coherent project and may have
interdependencies. The Elvish language itself only deals with modules; the
concept of package is a matter of how to organize modules.

Like the `go` command, Elvish does **not** have a central registry of
packages. A package is simply identified by the URL of its code repository,
e.g. [github.com/elves/sample-pkg](https://github.com/elves/sample-pkg). To
install the package, one simply uses the following:

```elvish
use epm
epm:install github.com/elves/sample-pkg
```

`epm` knows out-of-the-box how to manage packages hosted in GitHub,
BitBucket and GitLab, and requires the `git` command to be
available. It can also copy files via `git` or `rsync` from arbitrary
locations (see [Custom package domains](#custom-package-domains) for
details).

Once installed, modules in this package can be imported with `use
github.com/elves/sample-pkg/...`. This package has a module named
`sample-mod` containing a function `sample-fn`, and can be used like
this:

```elvish-transcript
~> use github.com/elves/sample-pkg/sample-mod
~> sample-mod:sample-fn
This is a sample function in a sample module in a sample package
```

The next section describes functions in the `epm` module, using the
same notation as the [doc for the builtin
module](builtin.html#usage-notation).

# Functions

## install

```elvish
epm:install &silent-if-installed=$false $pkg...
```

Install the named packages. By default, if a package is already
installed, a message will be shown. This can be disabled by passing
`&silent-if-installed=$true`, so that already-installed packages are
silently ignored.

## installed

```elvish
epm:installed
```

Return an array with all installed packages. `epm:list` can be used as
an alias for `epm:installed`.

## is-installed

```elvish
epm:is-installed $pkg
```

Returns a boolean value indicating whether the given package is
installed.

## metadata

```elvish
epm:metadata $pkg
```

Returns a hash containing the metadata for the given package. Metadata
for a package includes the following base attributes:

- `name`: name of the package
- `installed`: a boolean indicating whether the package is currently installed
- `method`: method by which it was installed (`git` or `rsync`)
- `src`: source URL of the package
- `dst`: where the package is (or would be) installed. Note that this attribute is returned even if `installed` is `$false`.

Additionally, packages can define arbitrary metadata attributes in a
file called `metadata.json` in their top directory. The following attributes are recommended:

- `description`: a human-readable description of the package
- `maintainers`: an array containing the package maintainers, in `Name <email>` format.
- `homepage`: URL of the homepage for the package, if it has one.

## query

```elvish
epm:query $pkg
```

Pretty print the available metadata of the given package.

## uninstall

```elvish
epm:install $pkg...
```

Uninstall named packages.

## upgrade

```elvish
epm:upgrade $pkg...
```

Upgrade named packages. If no package name is given, upgrade all installed
packages.

# Custom package domains

Package names in `epm` have the following structure:
`domain/path`. The `domain` is usually the hostname from where the
package is to be fetched, such as `github.com`. The `path` can have
one or more components separated by slashes. Usually, the full name of
the package corresponds with the URL from where it can be fetched. For
example, the package hosted at https://github.com/elves/sample-pkg is
identified as `github.com/elves/sample-pkg`.

Packages are stored under `~/.elvish/lib/` in a path identical to
their name. For example, the package mentioned above is stored at
`~/.elvish/lib/github.com/elves/sample-pkg`.

Each domain must be configured with the following information:

* The method to use to fetch packages from the domain. The two
  supported methods are `git` and `rsync`.

* The number of directory levels under the domain directory in which
  the packages are found. For example, for `github.com` the number of
  levels is 2, since package paths have two levels
  (e.g. `elves/sample-pkg`). All packages from a given domain have the
  same number of levels.

* Depending on the method, other attributes are needed:

    - `git` needs a `protocol` attribute, which can be `https` or
      `http`, and determines how the URL is constructed.

    - `rsync` needs a `location` attribute, which must be a valid source
      directory recognized by the `rsync` command.

`epm` includes default domain configurations for `github.com`,
`gitlab.com` and `bitbucket.org`. These three domains share the same
configuration:

```json
{
   "method" : "git",
   "protocol" : "https",
   "levels" : "2"
}
```

You can define your own domain by creating a file named
`epm-domain.cfg` in the appropriate directory under
`~/.elvish/lib/`. For example, if you want to define an `elvish-dev`
domain which installs packages from your local `~/dev/elvish/`
directory, you must create the file
`~/.elvish/lib/elvish-dev/epm-domain.cfg` with the following JSON
content:

```json
{
   "method" : "rsync",
   "location" : "~/dev/elvish",
   "levels" : "1"
}
```

You can then install any directory under `~/dev/elvish/` as a
package. For example, if you have a directory
`~/dev/elvish/utilities/`, the following command will install it under
`~/.elvish/lib/elvish-dev/utilities`:

```elvish
epm:install elvish-dev/utilities
```

When you make any changes to your source directory, `epm:upgrade` will
synchronize those changes to `~/.elvish/lib`.
