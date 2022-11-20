package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	for _, goFile := range os.Args[1:] {
		bs, err := os.ReadFile(goFile)
		handleErr("read file:", err)

		var goLines, elvLines []string

		lines := strings.Split(string(bs), "\n")
		for i := 0; i < len(lines); i++ {
			if !strings.HasPrefix(lines[i], "//elvdoc:") {
				goLines = append(goLines, lines[i])
				continue
			}
			if len(elvLines) > 0 {
				elvLines = append(elvLines, "")
			}
			elvLines = append(elvLines, "#"+lines[i][2:])
			i++
			for i < len(lines) && strings.HasPrefix(lines[i], "//") {
				elvLines = append(elvLines, "#"+lines[i][2:])
				i++
			}
			i--
		}

		os.WriteFile(goFile, []byte(strings.Join(goLines, "\n")), 0o644)
		if len(elvLines) > 0 {
			elvFile := goFile[:len(goFile)-len(filepath.Ext(goFile))] + ".d.elv"
			elvLines = append(elvLines, "")
			os.WriteFile(elvFile, []byte(strings.Join(elvLines, "\n")), 0o644)
		}
	}
}

func handleErr(s string, err error) {
	if err != nil {
		log.Fatalln(s, err)
	}
}
