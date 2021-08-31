import glob
import os
import sys
import urllib.parse

import bs4

PRELUDE = """
DROP TABLE IF EXISTS searchIndex;
CREATE TABLE searchIndex(id INTEGER PRIMARY KEY, name TEXT, type TEXT, path TEXT);
CREATE UNIQUE INDEX anchor ON searchIndex (name, type, path);
""".strip()

def main(args):
    if len(args) != 2:
        print('Usage: dsindex dir')
        sys.exit(1)
    os.chdir(args[1])

    print(PRELUDE)

    for filename in glob.glob('*.html'):
        with open(filename) as f:
            soup = bs4.BeautifulSoup(f, 'html.parser')
        anchors = soup.find_all('a', class_='dashAnchor')
        for anchor in anchors:
            name = anchor['name']
            entry_type, symbol = name.split('/')[-2:]
            symbol = urllib.parse.unquote(symbol)
            print(
                'INSERT OR IGNORE INTO searchIndex(name, type, path) VALUES '
                ' ("%s", "%s", "%s#%s");' % (symbol, entry_type, filename, name))

if __name__ == '__main__':
    main(sys.argv)
