if (eq $E:UPLOAD_TOKEN '') {
    echo 'UPLOAD_TOKEN must be set'
    exit 2
}

version = (git describe --tags --always)
fname-prefix = elvish
if (not-eq $E:TRAVIS_TAG '') {
    fname-prefix = elvish-$E:TRAVIS_TAG
}

fn build [os arch]{
    fname = $fname-prefix'-'$os'-'$arch
    echo 'Going to build '$fname
    go build -ldflags "-X main.Version="$version -o $fname
    tar cfz $fname.tar.gz $fname
    curl https://ul.elvish.io/ -F name=$fname.tar.gz -F token=$E:UPLOAD_TOKEN -F file=@$fname
    echo 'Built and uploaded '$fname
}

build darwin amd64
for arch [386 amd64 arm64] {
    build linux $arch
}
