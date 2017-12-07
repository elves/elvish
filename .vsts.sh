mkdir ~/bin
export PATH=$HOME/bin:$PATH
curl -sL -o ~/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x ~/bin/gimme
eval "$(gimme 1.9)"
go test ./...
