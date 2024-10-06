use etk
use str

var temperature = {|ctx:|
  ctx:subcomp celsius = (etk:with-init $etk:textarea [&prompt=(styled 'Celsius: ')])
  ctx:subcomp fahrenheit = (etk:with-init $etk:textarea [&prompt=(styled 'Fahrenheit: ')])
  ctx:state focus = (num 0)

  put [
    &view=(ctx:vbox celsius fahrenheit &focus=$ctx:state[focus])
    &react={|e|
      if (eq $e (etk:-key-event Tab)) {
        set ctx:state[focus] = (- 1 $ctx:state[focus])
        put $etk:consumed
      } elif (eq $e (etk:-key-event 'Ctrl-[')) {
        put $etk:finish
      } else {
        fn update {|from to conv|
          if (eq $etk:consumed (ctx:pass $from $e)) {
            try {
              $conv $ctx:state[$from][buffer][content] |
                var f = (printf '%.2f' (one))
              set ctx:state[$to][buffer] = [&content=$f &dot=(count $f)]
            } catch e {
            }
            put $etk:consumed
          } else {
            put $etk:unused
          }
        }
        if (== $ctx:state[focus] 0) {
          update celsius fahrenheit {|c| + 32 (* 9/5 $c)}
        } else {
          update fahrenheit celsius {|f| * 5/9 (- $f 32)}
        }
      }
    }
  ]
}

etk:run $temperature
