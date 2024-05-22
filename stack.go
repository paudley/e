// Copyright (C) 2022, 2023 by Blackcat Informatics® Inc.

package e

import (
	"fmt"
	"regexp"
	"runtime"
)

var (
	filestripRe  = regexp.MustCompile(`.*/`)
	funcfilterRe = regexp.MustCompile(`^(blackcat.ca/)?(?:db|output\.Wrap|cache\...Map.\.Clear)|^runtime`)
)

func CallLocation() (string, CallFrame) {
	backtrace, frames := FilteredStack()
	if len(backtrace) < 1 || len(frames) < 1 {
		return "(none)", CallFrame{File: "unknown", Line: 0, Function: "unknown"}
	}

	return backtrace[0], frames[0]
}

type CallFrame struct {
	File     string
	Line     int
	Function string
}

const maxCallerLength = 20

// FilteredStack returns a set of strings representing the interseting portions of the stack.
//
// nolint: funlen,cyclop
func FilteredStack() ([]string, []CallFrame) {
	pcs := make([]uintptr, maxCallerLength)
	n := runtime.Callers(1, pcs)
	stack := pcs[:n]
	frames := runtime.CallersFrames(stack)
	_, more := frames.Next() // special frame.
	ret := []string{}
	ret2 := []CallFrame{}

	for more {
		var frame runtime.Frame
		frame, more = frames.Next()

		if funcfilterRe.MatchString(frame.Function) {
			continue
		}

		switch frame.Function {
		case "github.com/paudley/e.CallLocation":
		case "github.com/paudley/e.New[...]":
		case "github.com/paudley/e.Full[...]":
		case "github.com/paudley/e.NewWithContext[...]":
		case "github.com/paudley/e.NewWithVals[...]":
		case "github.com/paudley/e.Wrap":
		case "github.com/paudley/e.WrapError[...]":
		case "github.com/paudley/e.WrapErrorMsg[...]":
		case "github.com/paudley/e.WrapErrorCtx[...]":
		case "github.com/paudley/e.newPathElement":
		case "github.com/paudley/e.WrapWithVals":
		case "github.com/paudley/e.WrapErr":
		case "blackcat.ca/fin.WithAppTx.func1.1":
		case "blackcat.ca/fin.WithAppTx.func1":
		case "blackcat.ca/app.(*EnhLogger).E":
		case "blackcat.ca/app.fatalDump":
		case "blackcat.ca/app.(*EnhLogger).Fatal":
		case "blackcat.ca/app.(*EnhLogger).Error":
		case "blackcat.ca/app.(*EnhLogger).Info":
		case "blackcat.ca/app.(*EnhLogger).Warn":
		case "blackcat.ca/app.(*EnhLogger).Debug":
		case "blackcat.ca/app.(*Logger).Fatal":
		case "blackcat.ca/app.(*Logger).Error":
		case "blackcat.ca/app.(*Logger).Info":
		case "blackcat.ca/app.(*Logger).Warn":
		case "blackcat.ca/app.(*Logger).Debug":
		case "github.com/paudley/e_test.cBad":
		default:
			file := frame.File
			file = filestripRe.ReplaceAllString(file, ``)
			ret = append(ret, fmt.Sprintf("%s@%s:%d", frame.Function, file, frame.Line))
			ret2 = append(ret2, CallFrame{File: file, Line: frame.Line, Function: frame.Function})
		}
	}

	return ret, ret2
}
