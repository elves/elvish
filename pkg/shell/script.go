package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

// ScriptConfig keeps configuration for the script mode.
type ScriptConfig struct {
	SpawnDaemon bool
	Paths       Paths

	Cmd         bool
	CompileOnly bool
	JSON        bool
}

// Script executes a shell script.
func Script(fds [3]*os.File, args []string, cfg *ScriptConfig) int {
	ev, cleanup := setupShell(fds, cfg.Paths, cfg.SpawnDaemon)
	defer cleanup()

	arg0 := args[0]
	ev.SetArgs(args[1:])

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
		parseErr, compileErr := ev.Check(src, fds[2])
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
		err := evalInTTY(ev, fds, src)
		if err != nil {
			diag.ShowError(fds[2], err)
			return 2
		}
	}

	return 0
}

var errSourceNotUTF8 = errors.New("source is not UTF-8")

func readFileUTF8(fname string) (string, error) {
	bytes, err := ioutil.ReadFile(fname)
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
func errorsToJSON(parseErr *parse.Error, compileErr *diag.Error) []byte {
	var converted []errorInJSON
	if parseErr != nil {
		for _, e := range parseErr.Entries {
			converted = append(converted,
				errorInJSON{e.Context.Name, e.Context.From, e.Context.To, e.Message})
		}
	}
	if compileErr != nil {
		converted = append(converted,
			errorInJSON{compileErr.Context.Name,
				compileErr.Context.From, compileErr.Context.To, compileErr.Message})
	}

	jsonError, errMarshal := json.Marshal(converted)
	if errMarshal != nil {
		return []byte(`[{"message":"Unable to convert the errors to JSON"}]`)
	}
	return jsonError
}
