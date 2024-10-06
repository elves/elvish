use etk

fn counter {|ctx:|
  ctx:state value = 0
  put [
    &view=(etk:-text-view $ctx:state[value] &dot-before=1)
    &react={|ev|
      if (eq $ev (etk:-key-event Enter)) {
        set ctx:state[value] = (+ 1 $ctx:state[value])
        put $etk:consumed
      } elif (eq $ev (etk:-key-event 'Ctrl-[')) {
        put $etk:finish
      } else {
        put $etk:unused
      }
    }
  ]
}

etk:run $counter~
