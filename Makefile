EXE := das

main:
	go build -o $(EXE) ./main

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: edit/tty/z-types.go

.PHONY: main pre-commit
.DEFAULT: main
