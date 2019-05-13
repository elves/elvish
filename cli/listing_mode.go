package cli

import "github.com/elves/elvish/cli/listing"

// ListingUp moves the listing selection up.
func ListingUp(ev KeyEvent) {
	ev.App().Listing.MutateStates((*listing.State).Up)
}

// ListingDown moves the listing selection down.
func ListingDown(ev KeyEvent) {
	ev.App().Listing.MutateStates((*listing.State).Down)
}

// ListingUpCycle moves the listing selection up, wrapping.
func ListingUpCycle(ev KeyEvent) {
	ev.App().Listing.MutateStates((*listing.State).UpCycle)
}

// ListingDownCycle moves the listing selection down, wrapping.
func ListingDownCycle(ev KeyEvent) {
	ev.App().Listing.MutateStates((*listing.State).DownCycle)
}

// ListingToggleFiltering toggles the filtering state of the listing mode.
func ListingToggleFiltering(ev KeyEvent) {
	ev.App().Listing.MutateStates((*listing.State).ToggleFiltering)
}

// ListingAccept accepts the current item of the listing.
func ListingAccept(ev KeyEvent) {
	ev.App().Listing.AcceptItem(ev.State())
}

// ListingAcceptClose accepts the current item of the listing and closes the
// listing.
func ListingAcceptClose(ev KeyEvent) {
	ev.App().Listing.AcceptItemAndClose(ev.State())
}

// ListingDefault invokes the default key handler of the listing mode.
func ListingDefault(ev KeyEvent) {
	ev.App().Listing.DefaultHandler(ev.State())
}
