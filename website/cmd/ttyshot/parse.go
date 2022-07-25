package main

import (
	"bytes"
	"errors"
	"regexp"
	"time"
)

// Operations for driving a demo ttyshot.
const (
	opEnter          = iota // enable implicit Enter key and send an Enter key
	opNoEnter               // inhibit implicit Enter key
	opTrimEmptyLines        // trim trailing empty lines -- can occur anywhere in the spec
	opUp                    // send Up arrow sequence
	opDown                  // send Down arrow sequence
	opRight                 // send Right arrow sequence
	opLeft                  // send Left arrow sequence
	opText                  // send the provided text, optionally followed by Enter
	opAlt                   // send an alt sequence
	opCtrl                  // send a control character
	opSleep                 // sleep for the specified duration
	opWaitForPrompt         // wait for the expected "<n>" (command number) in the next prompt
	opWaitForRegexp         // wait for sequence of bytes matching the regexp
	opWaitForString         // wait for the literal sequence of bytes
)

type demoOp struct {
	what int
	val  any
}

func parseSpec(content []byte) ([]demoOp, error) {
	lines := bytes.Split(content, []byte{'\n'})
	ops := make([]demoOp, 1, len(lines)+2)
	ops[0] = demoOp{opWaitForPrompt, nil}

	for _, line := range lines {
		if len(line) == 0 {
			continue // ignore empty lines
		}
		if bytes.HasPrefix(line, []byte("//")) {
			directive, err := parseDirective(line[2:])
			if err != nil {
				return ops, err
			}
			ops = append(ops, directive)
		} else {
			ops = append(ops, demoOp{opText, line})
		}
	}

	return ops, nil
}

func parseDirective(directive []byte) (demoOp, error) {
	if bytes.HasPrefix(directive, []byte("sleep ")) {
		duration, err := time.ParseDuration(string(directive[6:]) + "s")
		if err != nil {
			return demoOp{}, err
		}
		return demoOp{opSleep, duration}, nil
	}

	if bytes.Equal(directive, []byte("no-enter")) {
		return demoOp{opNoEnter, nil}, nil
	}

	if bytes.Equal(directive, []byte("trim-empty")) {
		return demoOp{opTrimEmptyLines, nil}, nil
	}

	if bytes.Equal(directive, []byte("enter")) {
		return demoOp{opEnter, nil}, nil
	}

	// Tab is frequently used so it's useful to support it as a directive rather than requiring
	// `//ctrl I`.
	if bytes.Equal(directive, []byte("tab")) {
		return demoOp{opCtrl, byte('I')}, nil
	}

	if bytes.HasPrefix(directive, []byte("ctrl ")) {
		if len(directive) != 6 {
			return demoOp{}, errors.New("invalid ctrl directive: " + string(directive))
		}
		return demoOp{opCtrl, directive[5]}, nil
	}

	if bytes.HasPrefix(directive, []byte("alt ")) {
		if len(directive) != 5 {
			return demoOp{}, errors.New("invalid alt directive: " + string(directive))
		}
		return demoOp{opAlt, directive[4]}, nil
	}

	if bytes.Equal(directive, []byte("prompt")) {
		return demoOp{opWaitForPrompt, nil}, nil
	}

	if bytes.Equal(directive, []byte("up")) {
		return demoOp{opUp, nil}, nil
	}

	if bytes.Equal(directive, []byte("down")) {
		return demoOp{opDown, nil}, nil
	}

	if bytes.Equal(directive, []byte("right")) {
		return demoOp{opRight, nil}, nil
	}

	if bytes.Equal(directive, []byte("left")) {
		return demoOp{opLeft, nil}, nil
	}

	if bytes.HasPrefix(directive, []byte("wait-for-re ")) {
		re, err := regexp.Compile(string(directive[12:]))
		if err != nil {
			return demoOp{}, errors.New("invalid wait-for-re value: " + string(directive[12:]))
		}
		return demoOp{opWaitForRegexp, re}, nil
	}

	if bytes.HasPrefix(directive, []byte("wait-for-str ")) {
		return demoOp{opWaitForString, directive[13:]}, nil
	}

	return demoOp{}, errors.New("unrecognized directive: " + string(directive))
}
