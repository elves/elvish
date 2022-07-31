use math
//prompt
var stateColors = [&open=fg-bright-green &closed=fg-red]
//prompt
fn colored {|state|
    put (styled {$state"  "}[..6] $stateColors[$state])
}
//prompt
var url = "https://api.github.com/repos/elves/elvish/issues?state=all&sort=updated&per_page=5"
//prompt
curl -s $url | from-json | all (one) |
each {|issue|
    var id = (exact-num $issue[number])
    var t = $issue[title]
    var title = $t[..(math:min 45 (count $t))]
    var state = $issue[state]
    echo (colored $state) $id $title
}
//prompt
