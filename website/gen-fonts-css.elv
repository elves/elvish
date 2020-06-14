# Generates fonts.css by doing the following:
#
# 1. Download Source Serif Pro and Fira Mono
#
# 2. Downsize them by only keeping the Latin glyphs
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
subset=…’“”

mkdir -p _fonts_tmp
pwd=_fonts_tmp {
  @ssp-files-base = SourceSerifPro-{Regular It Semibold SemiboldIt}
  for base $ssp-files-base {
    curl -C - -L -o $base.otf -s https://github.com/adobe-fonts/source-serif-pro/raw/release/OTF/$base.otf
  }
  @fm-files-base = FiraMono-{Regular Bold}
  for base $fm-files-base {
    curl -C - -L -o $base.otf -s https://github.com/mozilla/Fira/raw/master/otf/$base.otf
  }
  for base [$@ssp-files-base $@fm-files-base] {
    fonttools subset $base.otf --unicodes=00-7f --text=$subset
    fonttools ttLib.woff2 compress -o $base.subset.woff2 $base.subset.otf
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

font-face 'Source Serif Pro' 400 normal SourceSerifPro-Regular
font-face 'Source Serif Pro' 400 italic SourceSerifPro-It
font-face 'Source Serif Pro' 600 normal SourceSerifPro-Semibold
font-face 'Source Serif Pro' 600 italic SourceSerifPro-SemiboldIt

font-face 'Fira Mono' 400 normal FiraMono-Regular
font-face 'Fira Mono' 600 normal FiraMono-Bold
