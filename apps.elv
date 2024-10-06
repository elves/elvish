use etk

# mkdir prompt - what can be done today:
var w
set w = (edit:new-codearea [&prompt=(styled 'mkdir:' inverse)' ' &on-submit={ mkdir (edit:get-state $w)[buffer][content] }])
edit:push-addon $w

# mkdir prompt - slightly cleaned up version
edit:push-addon (etk:new-codearea [&prompt='mkdir: ' &on-submit={|s| mkdir $s[buffer][content] }])

# mkdir prompt - state management version
var dirname
edit:push-addon {
  etk:textbox [&prompt='mkdir: ' &on-submit={ mkdir $dirname }] ^
              [&buffer=[&content=(bind dirname)]]
}

# Temperature conversion
var c = ''
etk:run-app {
  var f = (/ (- $c 32) 1.8)
  etk:vbox [&children=[
    (etk:textbox [&prompt='input: '] [&buffer=[&content=(bind c)]])
    (etk:label $c' ℉ = '$f' ℃')
  ]]
}

# Elvish configuration helper
var tasks = [
  [&name='Use readline binding'
   &detail='Readline binding enables keys like Ctrl-N, Ctrl-F'
   &eval-code=''
   &rc-code='use readline-binding']

  [&name='Install Carapace'
   &detail='Carapace provides completions.'
   &eval-code='brew install carapace'
   &rc-code='eval (carapace init elvish)']
]

fn execute-task {|task|
  eval $task[eval-code]
  eval $task[rc-code]
  echo $task[rc-code] >> $runtime:rc-file
}

var i = (num 0)
etk:run-app {
  etk:hbox [&children=[
    (etk:list [&items=$tasks &display={|t| put $t[name]} &on-submit=$execute-task~] ^
              [&selected=(bind i)])
    (etk:label $tasks[i][detail])
  ]]
}

# Markdown-driven presentation
var filename = 'a.md'
var @slides = (slurp < $filename |
               re:split '\n {0,3}((?:-[ \t]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})\n' (one))

var i = (num 0)
etk:run-app {
  etk:vbox [
    &binding=[&Left={|_| set i = (- $i 1) } &Right={|_| set i = (+ $i 1) }]
    &children=[
      (etk:label $slides[i])
      (etk:label (+ 1 $i)/(count $slides))
    ]
  ]
}
