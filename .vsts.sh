curl -sL -o gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x gimme
eval "$(./gimme 1.9)"

export GOPATH=$HOME/go
SRCDIR=$GOPATH/src/github.com/elves/elvish
mkdir -p $(dirname $SRCDIR)
cp -r $PWD $SRCDIR
cd $SRCDIR
make test
