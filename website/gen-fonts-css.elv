# Generates fonts.css by doing the following:
#
# 1. Download Source Serif and Fira Mono
#
# 2. Downsize them by only keeping the Latin glyphs (and a few more)
#
# 3. Embed into the CSS file as base64
#
# External dependencies:
#   base64: for encoding to base64
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
    fonttools subset $base.otf --unicodes=00-7f,386 --text=$subset --layout-features-=dnom,frac,locl,numr --name-IDs= --flavor=woff2
  }
}

fn font-face [family weight style file]{
  echo "@font-face {
    font-family: "$family";
    font-weight: "$weight";
    font-style: "$style";
    font-strecth: normal;
    src: url('data:font/woff2;charset=utf-8;base64,"(base64 -w0 _fonts_tmp/$file.subset.woff2 | slurp)"') format('woff2');
}"
}

font-face 'Source Serif' 400 normal SourceSerif4-Regular
font-face 'Source Serif' 400 italic SourceSerif4-It
font-face 'Source Serif' 600 normal SourceSerif4-Semibold
font-face 'Source Serif' 600 italic SourceSerif4-SemiboldIt

font-face 'Fira Mono' 400 normal FiraMono-Regular
font-face 'Fira Mono' 600 normal FiraMono-Bold
