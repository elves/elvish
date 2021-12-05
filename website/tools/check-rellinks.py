import dataclasses
import glob
import os
import os.path
import sys
import urllib.parse

import bs4

@dataclasses.dataclass
class Link:
    href: str
    parsed: urllib.parse.ParseResult

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
        links = [Link(href=e['href'], parsed=urllib.parse.urlparse(e['href']))
                 for e in soup.find_all('a', href=True)]
        rellinks[filename] = [link for link in links if link.parsed.scheme == '']
        targets[filename] = [e['id'] for e in soup.find_all(id=True)]

    def check(path, fragment):
        if path.endswith('.atom') and fragment == '':
            return True
        return path in targets and (fragment == '' or fragment in targets[path])

    has_broken = False

    for filename in rellinks:
        if filename.endswith('-release-notes.html'):
            continue
        dirname = os.path.dirname(filename)
        broken_links = []
        for link in rellinks[filename]:
            path = link.parsed.path
            if path == '':
                path = filename
            else:
                if os.path.splitext(path)[1] == '':
                    path += '/index.html'
                if path.startswith('/'):
                    path = path.lstrip('/')
                else:
                    path = os.path.normpath(os.path.join(dirname, path))
            if not check(path, link.parsed.fragment):
                broken_links.append(link.href)
        if broken_links:
            if not has_broken:
                print('Found broken links:')
                has_broken = True
            print(filename)
            for link in broken_links:
                print(f'    {link}')
    if has_broken:
        sys.exit(1)

if __name__ == '__main__':
    main(sys.argv)
