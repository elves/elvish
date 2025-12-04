use ./epm

epm:install &silent-if-installed github.com/giancosta86/ethereal@v1
epm:install &silent-if-installed github.com/giancosta86/velvet@v1

use github.com/giancosta86/velvet/v1/main velvet

velvet:velvet &must-pass