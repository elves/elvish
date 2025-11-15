use os
use path

fn create-command { |source-map|
  var sources-by-dest = [&]

  fn checkout { |reference|
    if (not (has-key $sources-by-dest $pwd)) {
      printf 'Fake Git: the directory "%s" was not cloned via this command instance!' $pwd |
        fail (all)
    }

    var source = $sources-by-dest[$pwd]

    var repository-map = $source-map[$source]

    if (not (has-key $repository-map $reference)) {
      printf 'Fake Git: missing reference "%s" in repository at source "%s"' $reference $source |
        fail (all)
    }

    put *[nomatch-ok] | each { |pwd-entry|
      os:remove-all $pwd-entry
    }

    var reference-files = $repository-map[$reference]

    keys $reference-files | each { |entry-path|
      var entry-parent = (path:dir $entry-path)

      os:mkdir-all $entry-parent

      var file-content = $reference-files[$entry-path]

      print $file-content > $entry-path
    }
  }

  fn clone { |source dest|
    if (not (has-key $source-map $source)) {
      printf 'Fake Git: missing source "%s" in source map' $source |
        fail (all)
    }

    set sources-by-dest = (
      assoc $sources-by-dest (path:abs $dest) $source
    )

    os:mkdir-all $dest

    tmp pwd = $dest

    checkout main
  }

  var commands = [
    &clone=$clone~
    &checkout=$checkout~
  ]

  fn fake-git { |command @arguments|
    if (not (has-key $commands $command)) {
      printf 'Fake Git: unsupported "%s" command' $command |
        fail (all)
    }

    $commands[$command] $@arguments
  }

  put $fake-git~
}