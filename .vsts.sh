mkdir ~/bin
export PATH=$HOME/bin:$PATH

curl -sL -o ~/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x ~/bin/gimme
eval "$(gimme 1.9)"

mkdir ~/go
export GOPATH=$HOME/go
go get -d github.com/elves/elvish
cd ~/go/github.com/elves/elvish
make test
