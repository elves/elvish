// Package daemon implements a service for mediating access to the data store,
// and its client.
//
// Most RPCs exposed by the service correspond to the methods of Store in the
// store package and are not documented here.
package daemon

import "github.com/elves/elvish/util"

var logger = util.GetLogger("[daemon] ")
