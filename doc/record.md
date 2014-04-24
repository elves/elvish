# Record mode

Hit ^R to start and stop recording:

```
~> ^R
RECORD 1 ~> var $a String
RECORD 2 ~> set $a = `lala`
RECORD 3 ~> echo $a
lala
RECORD 4 ~> ^R
save as (^C for discard): a.elvish
```

(Some style could be applied to `RECORD #`)
