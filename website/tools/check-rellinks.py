import dataclasses
import glob
import os
import os.path
import re
import sys
import urllib.parse

import bs4

@dataclasses.dataclass
class Link:
    href: str
    parsed_href: urllib.parse.ParseResult
    line_no: int
    col_no: int

def main(args):
    if len(args) != 2:
        print('Usage: check-rellinks dir')
        sys.exit(1)
    os.chdir(args[1])

    filenames = glob.glob('**/*.html', recursive=True)
    targets = {}
    rellinks = {}
    for filename in filenames:
        with open(filename) as f:
            soup = bs4.BeautifulSoup(f, 'html.parser')
        targets[filename] = [e['id'] for e in soup.find_all(id=True)]
        links = [Link(href=e['href'], parsed_href=urllib.parse.urlparse(e['href']),
                      line_no=e.sourceline, col_no=e.sourcepos+1) # e.sourcepos is 0-based
                 for e in soup.find_all('a', href=True)]
        rellinks[filename] = [link for link in links if link.parsed_href.scheme == '']

    def check(path, fragment):
        return path in targets and (fragment == '' or fragment in targets[path])

    has_broken = False

    for srcfile in rellinks:
        # Old release notes may refer to docs that have since been removed, so
        # don't check them.
        if srcfile.endswith('-release-notes.html'):
            continue
        for link in rellinks[srcfile]:
            # Special-case the relative link to the feed; all the other links
            # are to HTML files.
            if link.href == 'feed.atom':
                continue
            def report_that_link(does_what):
                print(f'{srcfile}:{link.line_no}:{link.col_no}: {link.href} {does_what}')
                nonlocal has_broken
                has_broken = True
            # Check that the path part has slashes in the expected places.
            # Slashes are significant for resolving relative links; relative
            # links with missing or superfluous slashes can still lead the
            # browser to the correct file, but will break the relative links on
            # the target page.
            #
            # For example, a relative link in ref/str.html to ./re.html/ will
            # navigate to the expected ref/re.html, but when on that page, a
            # relative link from ref/re.html to ./math.html will resolve
            # incorrectly to ref/re.html/math.html due to the trailing slash.
            #
            # These paths are less harmful in production since the HTTP server
            # of https://elv.sh knows how to redirect them to the correct URL,
            # but this still results in unnecessary roundtrips with the server
            # and is best avoided altogether.
            path = link.parsed_href.path
            # Duplicate slashes like ..//ref/
            if '//' in path:
                report_that_link('has duplicate slashes')
                continue
            # Links relative to the root like /ref are valid on the website but
            # breaks local previewing.
            if path.startswith('/'):
                report_that_link('is relative to the root and breaks local previewing')
                continue
            if path.endswith('/'):
                # Paths ending with / should be directory links. Check that it's
                # not a file link with a superfluous trailing slash, like
                # ./re.html/.
                if path.endswith('.html/'):
                    report_that_link('has a superfluous trailing slash')
                    continue
            # Paths not ending with / should be either empty or a file link.
            # Check that it's not a directory link with a missing trailing
            # slash, like ../ref.
            elif path != '' and not path.endswith('.html'):
                report_that_link('lacks a trailing slash')
                continue

            # Now check if the link target is valid.
            if path == '':
                dstfile = srcfile
            else:
                if path.endswith('/'):
                    path += 'index.html'
                dstfile = os.path.normpath(os.path.join(os.path.dirname(srcfile), path))
            if dstfile not in targets:
                report_that_link(f'links to non-existing page {dstfile}')
                continue
            fragment = link.parsed_href.fragment
            if fragment != '' and fragment not in targets[dstfile]:
                report_that_link(f'links to non-existing target {fragment} in {dstfile}')
    if has_broken:
        sys.exit(1)

if __name__ == '__main__':
    main(sys.argv)
