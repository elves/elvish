use os
use path
use ./fake-git

var valid-fake-git~ = (
  fake-git:create-command [
    &'<some url>'=[
      &main=[
        &'alpha.txt'='This is a sample test'
        &'beta/gamma/delta.txt'='This is another test!'
      ]

      &secondary=[
        &'alpha.txt'='This is another copy of alpha'
        &'omega.txt'='This is Omega'
        &'sigma/tau.txt'='This is Tau'
      ]

      &empty=[&]
    ]
  ]
)

fn -with-temp-directory { |block|
  var temp-dir = (os:temp-dir)
  defer {
    cd (path:dir $temp-dir)
    os:remove-all $temp-dir
  }

  $block $temp-dir
}

>> 'Fake Git command' {
  >> 'passing an unknown command' {
    >> 'should fail' {
      throws {
        valid-fake-git DODO
      } |
        get-fail-content |
        should-be 'Fake Git: unsupported "DODO" command'
    }
  }

  >> 'cloning' {
    >> 'when the source map does not include the given source' {
      >> 'should fail' {
        var fake-git~ = (fake-git:create-command [&])

        throws {
          fake-git clone '<some url>' (os:temp-dir)
        } |
          get-fail-content |
          should-be 'Fake Git: missing source "<some url>" in source map'
      }
    }

    >> 'when the source has no main branch' {
      >> 'should fail' {
        var fake-git~ = (fake-git:create-command [
          &'<some url>'=[&]
        ])

        throws {
          fake-git clone '<some url>' (os:temp-dir)
        } |
          get-fail-content |
          should-be 'Fake Git: missing reference "main" in repository at source "<some url>"'
      }
    }

    >> 'when a source with main is requested' {
      >> 'should clone its files' {
        -with-temp-directory { |dest|
          valid-fake-git clone '<some url>' $dest

          slurp < (path:join $dest alpha.txt) |
            should-be 'This is a sample test'

          slurp < (path:join $dest beta gamma delta.txt) |
            should-be 'This is another test!'
        }
      }
    }
  }

  >> 'checkout' {
    >> 'when the branch was not declared' {
      >> 'should fail' {
        -with-temp-directory { |dest|
          valid-fake-git clone '<some url>' $dest

          cd $dest

          throws {
            valid-fake-git checkout UNDECLARED
          } |
            get-fail-content |
            should-be 'Fake Git: missing reference "UNDECLARED" in repository at source "<some url>"'
        }
      }
    }

    >> 'when the target directory is not a cloned repository' {
      >> 'should fail' {
        -with-temp-directory { |dest|
          cd $dest

          throws {
            valid-fake-git checkout secondary
          } |
            get-fail-content |
            should-be (printf 'Fake Git: the directory "%s" was not cloned via this command instance!' $dest)
        }
      }
    }

    >> 'when the branch was declared' {
      >> 'the target should contain only the branch files' {
        -with-temp-directory { |dest|
          valid-fake-git clone '<some url>' $dest

          cd $dest

          valid-fake-git checkout secondary

          slurp < (path:join $dest alpha.txt) |
            should-be 'This is another copy of alpha'

          slurp < (path:join $dest omega.txt) |
            should-be 'This is Omega'

          path:join $dest beta gamma delta.txt |
            os:is-regular (all) |
            should-be $false

          slurp < (path:join $dest sigma tau.txt) |
            should-be 'This is Tau'
        }
      }
    }

    >> 'when the branch is empty' {
      >> 'the target should contain no more files' {
        -with-temp-directory { |dest|
          valid-fake-git clone '<some url>' $dest

          cd $dest

          valid-fake-git checkout empty

          put *[nomatch-ok] |
            put [(all)] |
            should-be []
        }
      }
    }

    >> 'when cloning branches in sibling directories' {
      >> 'both directories should coexist' {
        -with-temp-directory { |temp-dir|
          var main-dir = (path:join $temp-dir A)
          valid-fake-git clone '<some url>' $main-dir

          var secondary-dir = (path:join $temp-dir B)
          valid-fake-git clone '<some url>' $secondary-dir
          cd $secondary-dir
          valid-fake-git checkout secondary

          path:join $main-dir beta gamma delta.txt |
            os:is-regular (all) |
            should-be $true

          path:join $secondary-dir sigma tau.txt |
            os:is-regular (all) |
            should-be $true
        }
      }
    }
  }
}