if (eq $E:UPLOAD_TOKEN '') {
    echo 'UPLOAD_TOKEN must be set'
    exit 2
}

version = (git describe --tags --always)
fname-suffix = ''
if (not-eq $E:TRAVIS_TAG '') {
  fname-suffix = -$E:TRAVIS_TAG
}

fn build [os arch]{
    bin = elvish-$os'-'$arch$fname-suffix
    archive = $bin.tar.gz
    echo 'Going to build '$bin
    E:GOOS=$os E:GOARCH=$arch go build -ldflags "-X main.Version="$version -o $bin
    if (eq $os windows) {
      archive = $bin.zip
      cp $bin $bin.exe
      zip $archive $bin.exe
    } else {
      tar cfz $archive $bin
    }
    curl https://ul.elvish.io/ -F name=$archive -F token=$E:UPLOAD_TOKEN -F file=@$archive
    echo 'Built '$bin' and uploaded '$archive
}

build darwin amd64
build windows amd64
for arch [386 amd64 arm64] {
    build linux $arch
}
