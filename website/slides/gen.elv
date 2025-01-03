#!/usr/bin/env elvish
use str
use flag

fn main { |&title=Presentation md html|
  var content = (go run src.elv.sh/website/cmd/md2html < $md | slurp)
  slurp < template.html |
    str:replace '$common-css' (cat ../reset.css ../sgr.css | slurp) (one) |
    str:replace '$title' $title (one) |
    str:replace '$content' $content (one) |
    print (one) > $html
}

flag:call $main~ $args
