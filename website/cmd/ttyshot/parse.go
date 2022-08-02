//go:build !windows && !plan9 && !js

package main

import (
	"errors"
	"regexp"
	"strings"
)

type opType int

// Operations for driving a demo ttyshot.
const (
	opEnter         opType = iota // enable implicit Enter key and send an Enter key
	opNoEnter                     // inhibit implicit Enter key
	opUp                          // send Up arrow sequence
	opDown                        // send Down arrow sequence
	opRight                       // send Right arrow sequence
	opLeft                        // send Left arrow sequence
	opText                        // send the provided text, optionally followed by Enter
	opAlt                         // send an alt sequence
	opCtrl                        // send a control character
	opWaitForPrompt               // wait for prompt marker
	opWaitForRegexp               // wait for sequence of bytes matching the regexp
	opWaitForString               // wait for the literal sequence of bytes
)

type op struct {
	typ opType
	val any
}

func parseSpec(content string) ([]op, error) {
	lines := strings.Split(content, "\n")
	ops := make([]op, 1, len(lines)+2)
	ops[0] = op{opWaitForPrompt, nil}

	for _, line := range lines {
		if len(line) == 0 {
			continue // ignore empty lines
		}
		if strings.HasPrefix(line, "//") {
			directive, err := parseDirective(line[2:])
			if err != nil {
				return ops, err
			}
			ops = append(ops, directive)
		} else {
			ops = append(ops, op{opText, line})
		}
	}

	return ops, nil
}

func parseDirective(directive string) (op, error) {
	if directive == "no-enter" {
		return op{opNoEnter, nil}, nil
	}

	if directive == "enter" {
		return op{opEnter, nil}, nil
	}

	// Tab is frequently used so it's useful to support it as a directive rather than requiring
	// `//ctrl I`.
	if directive == "tab" {
		return op{opCtrl, byte('I')}, nil
	}

	if strings.HasPrefix(directive, "ctrl ") {
		if len(directive) != 6 {
			return op{}, errors.New("invalid ctrl directive: " + string(directive))
		}
		return op{opCtrl, directive[5]}, nil
	}

	if strings.HasPrefix(directive, "alt ") {
		if len(directive) != 5 {
			return op{}, errors.New("invalid alt directive: " + string(directive))
		}
		return op{opAlt, directive[4]}, nil
	}

	if directive == "prompt" {
		return op{opWaitForPrompt, nil}, nil
	}

	if directive == "up" {
		return op{opUp, nil}, nil
	}

	if directive == "down" {
		return op{opDown, nil}, nil
	}

	if directive == "right" {
		return op{opRight, nil}, nil
	}

	if directive == "left" {
		return op{opLeft, nil}, nil
	}

	if strings.HasPrefix(directive, "wait-for-re ") {
		re, err := regexp.Compile(string(directive[12:]))
		if err != nil {
			return op{}, errors.New("invalid wait-for-re value: " + string(directive[12:]))
		}
		return op{opWaitForRegexp, re}, nil
	}

	if strings.HasPrefix(directive, "wait-for-str ") {
		return op{opWaitForString, directive[13:]}, nil
	}

	return op{}, errors.New("unrecognized directive: " + string(directive))
}
