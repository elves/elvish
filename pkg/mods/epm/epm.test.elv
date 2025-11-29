use os
use path
use str
use ./epm
use ./fake-git

set epm:git~ = (fake-git:create-command [
  &'https://github.com/giancosta86/epm-plus'=[
    &main=[
      &'epm-plus.elv'='#This is a placeholder for the actual source code'
    ]

    &'v1.0.0+test1'=[
      &'alpha.elv'='
        fn f {
          put 90
        }
      '
    ]

    &'v1.0.0+test2'=[
      &'metadata.json'='
      {
        "description": "Test package with dependencies",

        "maintainers": ["Gianluca Costa <gianluca@gianlucacosta.info>"],

        "homepage": "https://github.com/giancosta86/epm-plus",

        "dependencies": ["github.com/giancosta86/epm-plus@v1.0.0+test1"]
      }
      '

      &'alpha.elv'='
        use ../v1.0.0+test1/alpha v1

        fn f {
          / (v1:f) 18
        }
      '

      &'beta.elv'='
        use github.com/giancosta86/epm-plus/v1.0.0+test1/alpha v1
        use ./alpha

        fn g {
          + (v1:f) (alpha:f)
        }
      '
    ]

    &'v1.0.0'=[
      &'README.MD'='
        # This is just some placeholder file
      '
    ]
  ]
])

var within-epm-plus-install~ = (
  var epm-plus-root = (path:join $epm:managed-dir github.com giancosta86 epm-plus)

  var active-blocks = 0

  put { |&reference=$nil block|
    if (== $active-blocks 0) {
      os:remove-all $epm-plus-root
    }

    set active-blocks = (+ $active-blocks 1)

    defer {
      set active-blocks = (- $active-blocks 1)

      if (== $active-blocks 0) {
        os:remove-all $epm-plus-root
      }
    }

    var pkg

    if $reference {
      set pkg = github.com/giancosta86/epm-plus@$reference
    } else {
      set pkg = github.com/giancosta86/epm-plus
    }

    epm:install $pkg

    {
      tmp pwd = (epm:dest $pkg)
      $block
    }
  }
)

fn get-test-package-list {
  epm:list |
    keep-if { |entry| str:has-prefix $entry github.com/giancosta86/epm-plus } |
    order |
    put [(all)]
}

>> 'In epm' {
  os:remove-all (path:join $epm:managed-dir github.com giancosta86 epm-plus)

  >> 'splitting package name and version' {
    >> 'without version' {
      epm:-split-package-name-and-version 'github.com/giancosta86/epm-plus' |
        put [(all)] |
        should-be [
          'github.com/giancosta86/epm-plus'
          $nil
        ]
    }

    >> 'with version' {
      epm:-split-package-name-and-version 'github.com/giancosta86/epm-plus@v1.0.0+test1' |
        put [(all)] |
        should-be [
          'github.com/giancosta86/epm-plus'
          'v1.0.0+test1'
        ]
    }
  }

  >> 'computing the destination directory of a package' {
    >> 'without version' {
      epm:dest github.com/giancosta86/velvet |
        should-be (path:join $epm:managed-dir github.com giancosta86 velvet)
    }

    >> 'with version' {
      epm:dest github.com/giancosta86/velvet@v2 |
        should-be (path:join $epm:managed-dir github.com giancosta86 velvet v2)
    }
  }

  >> 'installing from a Git repository' {
    >> 'when requesting no specific reference' {
      >> 'should perform a basic clone from main' {
        within-epm-plus-install {
          os:is-regular epm-plus.elv |
            should-be $true
        }
      }
    }

    >> 'when requesting a standalone reference' {
      within-epm-plus-install &reference=v1.0.0+test1 {
        >> 'should clone that reference' {
          os:is-regular epm-plus.elv |
            should-be $false

          os:is-regular alpha.elv |
            should-be $true
        }

        >> 'should not include the reference in the metadata src field' {
          epm:metadata github.com/giancosta86/epm-plus@v1.0.0+test1 |
            put (all)[src] |
            should-be https://github.com/giancosta86/epm-plus
        }
      }
    }

    >> 'when installing a reference depending on another' {
      within-epm-plus-install &reference=v1.0.0+test2 {
        >> 'the 2 reference directories should coexist' {
          os:is-regular alpha.elv |
            should-be $true

          os:is-regular beta.elv |
            should-be $true

          var test1-dest = (epm:dest github.com/giancosta86/epm-plus@v1.0.0+test1)

          os:is-dir $test1-dest |
            should-be $true

          path:join $test1-dest alpha.elv |
            os:is-regular (all) |
            should-be $true
        }

        >> 'one reference scripts should be callable from the other' {
          use github.com/giancosta86/epm-plus/v1.0.0+test1/alpha v1

          use github.com/giancosta86/epm-plus/v1.0.0+test2/alpha
          use github.com/giancosta86/epm-plus/v1.0.0+test2/beta

          v1:f |
            should-be 90

          alpha:f |
            should-be 5

          beta:g |
            should-be 95
        }
      }
    }
  }

  >> 'getting all the dependencies from metadata' {
    >> 'should list both dependencies and dev dependencies' {
      var metadata = [
        &dependencies=[alpha beta gamma]
        &devDependencies=[delta epsilon]
      ]

      epm:-get-all-dependencies $metadata |
        should-be [alpha beta gamma delta epsilon]
    }
  }

  >> 'listing Git packages' {
    >> 'when only the base reference is installed' {
      >> 'should list just the package name' {
        within-epm-plus-install {
          get-test-package-list |
            should-be [
              github.com/giancosta86/epm-plus
            ]
        }
      }
    }

    >> 'when a single reference is installed' {
      >> 'should list the package name with the reference' {
        within-epm-plus-install &reference=v1.0.0+test1 {
          get-test-package-list |
            should-be [
              github.com/giancosta86/epm-plus@v1.0.0+test1
            ]
        }
      }
    }

    >> 'when a reference installs another as a dependency' {
      >> 'should list both full packages' {
        within-epm-plus-install &reference=v1.0.0+test2 {
          get-test-package-list |
            should-be [
              github.com/giancosta86/epm-plus@v1.0.0+test1
              github.com/giancosta86/epm-plus@v1.0.0+test2
            ]
        }
      }
    }

    >> 'when the base version and a reference are installed' {
      >> 'should list just the package name' {
        within-epm-plus-install {
          within-epm-plus-install &reference=v1.0.0+test1 {
            get-test-package-list |
              should-be [
                github.com/giancosta86/epm-plus
              ]
          }
        }
      }
    }
  }

  >> 'uninstalling a Git package' {
    >> 'when the Git reference is specified' {
      >> 'the other references should remain' {
        within-epm-plus-install &reference=v1.0.0+test2 {
          cd ..

          epm:uninstall github.com/giancosta86/epm-plus@v1.0.0+test2

          os:is-dir v1.0.0+test2 |
            should-be $false

          os:is-dir v1.0.0+test1 |
            should-be $true
        }
      }

      >> 'when removing the other reference, too' {
        >> 'the package should no more be listed' {
          within-epm-plus-install &reference=v1.0.0+test2 {
            cd ..

            epm:uninstall github.com/giancosta86/epm-plus@v1.0.0+test1
            epm:uninstall github.com/giancosta86/epm-plus@v1.0.0+test2

            get-test-package-list |
              should-be []
          }
        }
      }
    }

    >> 'when no Git reference is specified' {
      >> 'the entire package directory should be deleted' {
        within-epm-plus-install &reference=v1.0.0+test2 {
          cd (path:join $epm:managed-dir github.com giancosta86)

          epm:uninstall github.com/giancosta86/epm-plus

          os:is-dir epm-plus |
            should-be $false
        }
      }
    }
  }

  >> 'installing without passing packages' {
    var temp-dir = (os:temp-dir)

    cd $temp-dir

    defer {
      cd (path:dir $temp-dir)
      os:remove-all $temp-dir
    }

    >> 'when the metadata descriptor is missing' {
      >> 'should display an error' {
        epm:install |
          str:contains (all) 'You must specify at least one package.' |
            should-be $true
      }
    }

    >> 'when the metadata descriptor is present' {
      var metadata = [
        &description='Test package'
        &dependencies=['github.com/giancosta86/epm-plus@v1.0.0']
        &devDependencies=['github.com/giancosta86/epm-plus@v1.0.0+test1']
      ]

      var all-dependencies = (epm:-get-all-dependencies $metadata)

      put $metadata |
        to-json > metadata.json

      var pre-install-packages = [(epm:list)]

      all $metadata[dependencies] | each { |dependency|
        if (has-value $pre-install-packages $dependency) {
          fail 'Dependency '$dependency' already installed! The test would be pointless!'
        }
      }

      all $metadata[devDependencies] | each { |dev-dependency|
        if (has-value $pre-install-packages $dev-dependency) {
          fail 'Dev dependency '$dev-dependency' already installed! The test would be pointless!'
        }
      }

      var install-output = (epm:install | slurp)

      defer {
        all $metadata[dependencies] | each { |dependency|
          epm:uninstall $dependency
        }

        all $metadata[devDependencies] | each { |dev-dependency|
          epm:uninstall $dev-dependency
        }
      }

      var post-install-packages = [(epm:list)]

      >> 'should display no error' {
        put $install-output |
          str:contains (all) 'You must specify at least one package.' |
            should-be $false
      }

      >> 'should make the dependencies available' {
        all $metadata[dependencies] | each { |dependency|
          has-value $post-install-packages $dependency |
            should-be $true
        }
      }

      >> 'should make the dev dependencies available' {
        all $metadata[devDependencies] | each { |dev-dependency|
          has-value $post-install-packages $dev-dependency |
            should-be $true
        }
      }
    }
  }
}