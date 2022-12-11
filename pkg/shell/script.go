package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

// Configuration for the script mode.
type scriptCfg struct {
	Cmd         bool
	CompileOnly bool
	JSON        bool
}

// Executes a shell script.
func script(ev *eval.Evaler, fds [3]*os.File, args []string, cfg *scriptCfg) int {
	arg0 := args[0]
	ev.Args = vals.MakeListSlice(args[1:])

	var name, code string
	if cfg.Cmd {
		name = "code from -c"
		code = arg0
	} else {
		var err error
		name, err = filepath.Abs(arg0)
		if err != nil {
			fmt.Fprintf(fds[2],
				"cannot get full path of script %q: %v\n", arg0, err)
			return 2
		}
		code, err = readFileUTF8(name)
		if err != nil {
			fmt.Fprintf(fds[2], "cannot read script %q: %v\n", name, err)
			return 2
		}
	}

	src := parse.Source{Name: name, Code: code, IsFile: true}
	if cfg.CompileOnly {
		parseErr, _, compileErr := ev.Check(src, fds[2])
		if cfg.JSON {
			fmt.Fprintf(fds[1], "%s\n", errorsToJSON(parseErr, compileErr))
		} else {
			if parseErr != nil {
				diag.ShowError(fds[2], parseErr)
			}
			if compileErr != nil {
				diag.ShowError(fds[2], compileErr)
			}
		}
		if parseErr != nil || compileErr != nil {
			return 2
		}
	} else {
		err := evalInTTY(fds, ev, nil, src)
		if err != nil {
			diag.ShowError(fds[2], err)
			return 2
		}
	}

	return 0
}

var errSourceNotUTF8 = errors.New("source is not UTF-8")

func readFileUTF8(fname string) (string, error) {
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(bytes) {
		return "", errSourceNotUTF8
	}
	return string(bytes), nil
}

// An auxiliary struct for converting errors with diagnostics information to JSON.
type errorInJSON struct {
	FileName string `json:"fileName"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Message  string `json:"message"`
}

// Converts parse and compilation errors into JSON.
func errorsToJSON(parseErr, compileErr error) []byte {
	var converted []errorInJSON
	for _, e := range parse.UnpackErrors(parseErr) {
		converted = append(converted,
			errorInJSON{e.Context.Name, e.Context.From, e.Context.To, e.Message})
	}
	for _, e := range eval.UnpackCompilationErrors(compileErr) {
		converted = append(converted,
			errorInJSON{e.Context.Name, e.Context.From, e.Context.To, e.Message})
	}

	jsonError, errMarshal := json.Marshal(converted)
	if errMarshal != nil {
		return []byte(`[{"message":"Unable to convert the errors to JSON"}]`)
	}
	return jsonError
}
