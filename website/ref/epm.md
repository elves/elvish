<!-- toc -->

@module epm

# Introduction

The Elvish Package Manager (`epm`) is a module bundled with Elvish for managing
third-party packages.

In Elvish terminology, a **module** is a `.elv` file that can be imported with
the `use` command, while a **package** is a collection of modules that are
usually kept in the same repository as one coherent project and may have
interdependencies. The Elvish language itself only deals with modules; the
concept of package is a matter of how to organize modules.

Like the `go` command, Elvish does **not** have a central registry of packages.
A package is simply identified by the URL of its code repository, e.g.
[github.com/elves/sample-pkg](https://github.com/elves/sample-pkg). To install
the package, one simply uses the following:

```elvish
use epm
epm:install github.com/elves/sample-pkg
```

`epm` knows out-of-the-box how to manage packages hosted in GitHub, BitBucket
and GitLab, and requires the `git` command to be available. It can also copy
files via `git` or `rsync` from arbitrary locations (see
[Custom package domains](#custom-package-domains) for details).

Once installed, modules in this package can be imported with
`use github.com/elves/sample-pkg/...`. This package has a module named
`sample-mod` containing a function `sample-fn`, and can be used like this:

```elvish-transcript
~> use github.com/elves/sample-pkg/sample-mod
~> sample-mod:sample-fn
This is a sample function in a sample module in a sample package
```

# The `epm`-managed directory

Elvish searches for modules in
[multiple directories](command.html#module-search-directories), and `epm` only
manages one of them:

-   On UNIX, `epm` manages `$XDG_DATA_HOME/elvish/lib`, defaulting to
    `~/.local/share/elvish/lib` if `$XDG_DATA_HOME` is unset or empty;

-   On Windows, `epm` manages `%LocalAppData%\elvish\lib`.

This directory is called the `epm`-managed directory, and its path is available
as [`$epm:managed-dir`]().

# Custom package domains

Package names in `epm` have the following structure: `domain/path`. The `domain`
is usually the hostname from where the package is to be fetched, such as
`github.com`. The `path` can have one or more components separated by slashes.
Usually, the full name of the package corresponds with the URL from where it can
be fetched. For example, the package hosted at
https://github.com/elves/sample-pkg is identified as
`github.com/elves/sample-pkg`.

Packages are stored under the `epm`-managed directory in a path identical to
their name. For example, the package mentioned above is stored at
`$epm:managed-dir/github.com/elves/sample-pkg`.

Each domain must be configured with the following information:

-   The method to use to fetch packages from the domain. The two supported
    methods are `git` and `rsync`.

-   The number of directory levels under the domain directory in which the
    packages are found. For example, for `github.com` the number of levels is 2,
    since package paths have two levels (e.g. `elves/sample-pkg`). All packages
    from a given domain have the same number of levels.

-   Depending on the method, other attributes are needed:

    -   `git` needs a `protocol` attribute, which can be `https` or `http`, and
        determines how the URL is constructed.

    -   `rsync` needs a `location` attribute, which must be a valid source
        directory recognized by the `rsync` command.

`epm` includes default domain configurations for `github.com`, `gitlab.com` and
`bitbucket.org`. These three domains share the same configuration:

```json
{
    "method": "git",
    "protocol": "https",
    "levels": "2"
}
```

You can define your own domain by creating a file named `epm-domain.cfg` in the
appropriate directory under `$epm:managed-dir`. For example, if you want to
define an `elvish-dev` domain which installs packages from your local
`~/dev/elvish/` directory, you must create the file
`$epm:managed-dir/elvish-dev/epm-domain.cfg` with the following JSON content:

```json
{
    "method": "rsync",
    "location": "~/dev/elvish",
    "levels": "1"
}
```

You can then install any directory under `~/dev/elvish/` as a package. For
example, if you have a directory `~/dev/elvish/utilities/`, the following
command will install it under `$epm:managed-dir/elvish-dev/utilities`:

```elvish
epm:install elvish-dev/utilities
```

When you make any changes to your source directory, `epm:upgrade` will
synchronize those changes to `$epm:managed-dir`.
