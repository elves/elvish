import glob
import os
import sys
import sqlite3
import urllib.parse

import bs4

PRELUDE = """
DROP TABLE IF EXISTS searchIndex;
CREATE TABLE searchIndex(id INTEGER PRIMARY KEY, name TEXT, type TEXT, path TEXT);
CREATE UNIQUE INDEX anchor ON searchIndex (name, type, path);
""".strip()

def main(args):
    if len(args) != 3:
        print('Usage: mkdsidx.py dir-of-html-files output-db-file')
        sys.exit(1)
    html_dir, output_db = args[1:]

    sql_statements = [PRELUDE]

    for filename in glob.glob('*.html', root_dir=html_dir):
        with open(os.path.join(html_dir, filename)) as f:
            soup = bs4.BeautifulSoup(f, 'html.parser')
        anchors = soup.find_all('a', class_='dashAnchor')
        for anchor in anchors:
            name = anchor['name']
            entry_type, symbol = name.split('/')[-2:]
            symbol = urllib.parse.unquote(symbol)
            sql_statements.append(
                "INSERT OR IGNORE INTO searchIndex(name, type, path) VALUES "
                " ('%s', '%s', '%s#%s');" % (symbol, entry_type, filename, name))

    with sqlite3.connect(output_db) as conn:
        conn.cursor().executescript(''.join(sql_statements))

if __name__ == '__main__':
    main(sys.argv)
