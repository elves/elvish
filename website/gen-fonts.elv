# Download the fonts needed by the website and downsize them by subsetting.
#
# External dependencies:
#   curl: for downloading files
#   fonttools: for processing font files

# Subset of glyphs to include, other than ASCII. Discovered with:
#
# cat **.html | go run ./cmd/runefreq | sort -nr
var subset = …’“”

mkdir -p _fonts_tmp
pwd=_fonts_tmp {
  @ssp-files-base = SourceSerif4-{Regular It Semibold SemiboldIt}
  for base $ssp-files-base {
    curl -C - -L -o $base.otf -s https://github.com/adobe-fonts/source-serif/raw/release/OTF/$base.otf
  }
  @fm-files-base = FiraMono-{Regular Bold}
  for base $fm-files-base {
    curl -C - -L -o $base.otf -s https://github.com/mozilla/Fira/raw/master/otf/$base.otf
  }

  for base [$@ssp-files-base $@fm-files-base] {
    # For some reason I don't understand, without U+386, the space (U+20) in
    # Fira Mono will be more narrow than other glyphs, so we keep it.
    fonttools subset $base.otf --output-file=../fonts/$base.woff2 --flavor=woff2 --with-zopfli ^
      --unicodes=00-7f,386 --text=$subset --layout-features-=dnom,frac,locl,numr --name-IDs=
  }
}
