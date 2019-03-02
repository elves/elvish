package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/listing"
)

func initListing(ed editor, ev *eval.Evaler) (*listing.Mode, eval.Ns) {
	m := &listing.Mode{}
	ns := eval.Ns{}
	return m, ns
}
