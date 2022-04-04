package eval

import (
	"errors"
	"os"
)

// ErrNonExistentEnvVar is raised by the get-env command when the environment
// variable does not exist.
var ErrNonExistentEnvVar = errors.New("non-existent environment variable")

//elvdoc:fn set-env
//
// ```elvish
// set-env $name $value
// ```
//
// Sets an environment variable to the given value. Example:
//
// ```elvish-transcript
// ~> set-env X foobar
// ~> put $E:X
// ▶ foobar
// ```
//
// @cf get-env has-env unset-env

//elvdoc:fn unset-env
//
// ```elvish
// unset-env $name
// ```
//
// Unset an environment variable. Example:
//
// ```elvish-transcript
// ~> set-env X foo
// ~> unset-env X
// ~> has-env X
// ▶ $false
// ~> put $E:X
// ▶ ''
// ```
//
// @cf has-env get-env set-env

func init() {
	addBuiltinFns(map[string]any{
		"has-env":   hasEnv,
		"get-env":   getEnv,
		"set-env":   os.Setenv,
		"unset-env": os.Unsetenv,
	})
}

//elvdoc:fn has-env
//
// ```elvish
// has-env $name
// ```
//
// Test whether an environment variable exists. Examples:
//
// ```elvish-transcript
// ~> has-env PATH
// ▶ $true
// ~> has-env NO_SUCH_ENV
// ▶ $false
// ```
//
// @cf get-env set-env unset-env

func hasEnv(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

//elvdoc:fn get-env
//
// ```elvish
// get-env $name
// ```
//
// Gets the value of an environment variable. Throws an exception if the
// environment variable does not exist. Examples:
//
// ```elvish-transcript
// ~> get-env LANG
// ▶ zh_CN.UTF-8
// ~> get-env NO_SUCH_ENV
// Exception: non-existent environment variable
// [tty], line 1: get-env NO_SUCH_ENV
// ```
//
// @cf has-env set-env unset-env

func getEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrNonExistentEnvVar
	}
	return value, nil
}
