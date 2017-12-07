mkdir ~/bin ~/go
export PATH=$HOME/bin:$PATH

curl -sL -o ~/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x ~/bin/gimme
eval "$(gimme 1.9)"

export GOPATH=$HOME/go
go get github.com/elves/elvish
go test github.com/elves/elvish/...
