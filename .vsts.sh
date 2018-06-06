curl -sL -o gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x gimme
eval "$(./gimme 1.9)"

export GOPATH=$HOME/go
mkdir -p $GOPATH/src/github.com/elves
ln -s $PWD $GOPATH/src/github.com/elves/elvish
make test
