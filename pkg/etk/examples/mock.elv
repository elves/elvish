#!/usr/bin/env elvish
use flag

fn watch-mock {|fname|
 { echo; fswatch $fname } | each {|_|
   clear
   try { print (render-styledown (slurp < $fname)) } catch e { echo $e }
 }
}

flag:call $watch-mock~ $args
