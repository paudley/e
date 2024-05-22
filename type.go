// Copyright (C) 2022, 2023, 2024 by Blackcat Informatics® Inc.

package e

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// V wraps variables that you want to include in a error as extra data.
type V struct {
	K string
	I any
}

type (
	// Value is any variable that you want to add to an error frame.
	Value any
	// Values is a list of variables added to an error frame.
	Values []Value
)

// ValueFunc is for lamndas that capture local values in error frames.
type ValueFunc func() Values

type PathElement struct {
	FileName   string
	LineNumber int
	FuncName   string
	Msg        string
	ValFunc    ValueFunc
	values     Values
	valuesP    bool
}

type ErrorClass interface {
	What() string
	Area() string
	Number() uint32
}

// The unexported internal error pointer.  This is what get's passed
// around but is opaque by design to calling code.
type errorData struct {
	createdAt   time.Time
	originerror error
	// We make a copy of the ctx here, not to use in a context sense but so we can dump it as part of the error.
	//nolint: containedctx
	originContext  context.Context
	originContextP bool
	class          ErrorClass
	path           []PathElement
}

// The externally accessible error type.  Use this in your returns.
type Error = *errorData

// Resolve the set of values for this path element.
func (pe PathElement) Values() (vals Values) {
	if pe.ValFunc == nil {
		return Values{}
	}

	if pe.valuesP {
		return pe.values
	}

	defer func() {
		if panicErr := recover(); panicErr != nil {
			// panic'ed in ValFunc.  Not a good sign...
			msg := "unknown"
			switch errV := panicErr.(type) {
			case string:
				msg = errV

			case error:
				msg = errV.Error()
			}

			vals = Values{V{"panic", "PANIC in ValFunc: " + msg}}
		}
	}()

	vals = pe.values
	vals = append(vals, pe.ValFunc()...)

	return vals
}

func newPathElement(msg string) PathElement {
	_, pc := CallLocation()

	return PathElement{
		FileName:   pc.File,
		LineNumber: pc.Line,
		FuncName:   pc.Function,
		Msg:        msg,
	}
}

func New[T ErrorClass](msg string) Error {
	var ec T

	return &errorData{
		class:     ec,
		createdAt: time.Now(),
		path:      []PathElement{newPathElement(msg)},
	}
}

func NewWithContext[T ErrorClass](ctx context.Context, msg string) Error {
	e := New[T](msg)
	e.originContext = ctx
	e.originContextP = true

	return e
}

func NewWithVals[T ErrorClass](msg string, valFunc ValueFunc) Error {
	e := New[T](msg)
	e.path[0].ValFunc = valFunc

	return e
}

func Full[T ErrorClass](ctx context.Context, msg string, valFunc ValueFunc) Error {
	e := NewWithContext[T](ctx, msg)
	e.path[0].ValFunc = valFunc

	return e
}

func Wrap(errorToWrap Error, msg string) Error {
	if errorToWrap == nil {
		return New[UnknownError](msg)
	}

	errorToWrap.path = append(errorToWrap.path, newPathElement(msg))

	return errorToWrap
}

func WrapError[T ErrorClass](err error) Error {
	if err == nil {
		return New[NoError]("no error")
	}

	eData := New[T](err.Error())
	eData.originerror = err

	eData.path[0].ValFunc = func() Values {
		return Values{
			"wrapped_error",
			V{K: "err", I: err},
		}
	}

	return eData
}

func WrapErrorMsg[T ErrorClass](err error, msg string) Error {
	if err == nil {
		return New[NoError]("no error")
	}

	eData := New[T](err.Error())
	eData.path[0].ValFunc = func() Values {
		return Values{
			"wrapped_error",
			V{K: "err", I: err},
		}
	}
	eData.path = append(eData.path, newPathElement(msg))

	return eData
}

func (e Error) AddValues(valFunc ValueFunc) Error {
	valFuncOrig := e.path[len(e.path)-1].ValFunc
	e.path[len(e.path)-1].ValFunc = func() Values {
		return append(valFuncOrig(), valFunc()...)
	}

	return e
}

func (e Error) AddValue(key string, val any) Error {
	valFuncOrig := e.path[len(e.path)-1].ValFunc
	e.path[len(e.path)-1].ValFunc = func() Values {
		return append(valFuncOrig(), func() Values { return Values{V{K: key, I: val}} })
	}

	return e
}

func WrapErrorCtx[T ErrorClass](ctx context.Context, err error) Error {
	e := WrapError[T](err)
	e.originContext = ctx
	e.originContextP = true

	return e
}

func WrapWithVals[T ErrorClass](errorToWrap Error, msg string, valFunc ValueFunc) Error {
	if errorToWrap == nil {
		return New[T](msg)
	}

	errorToWrap.path = append(errorToWrap.path, newPathElement(msg))
	errorToWrap.path[len(errorToWrap.path)-1].ValFunc = valFunc

	return errorToWrap
}

// FullWrap wraps an Error, adds a message and values and possibly updates missing context information.
func FullWrap[T ErrorClass](ctx context.Context, errorToWrap Error, msg string, valFunc ValueFunc) Error {
	if !errorToWrap.originContextP {
		errorToWrap.originContext = ctx
		errorToWrap.originContextP = true
	}

	return WrapWithVals[T](errorToWrap, msg, valFunc)
}

func (e Error) Path() []PathElement {
	return e.path
}

func (e Error) OriginContext() *context.Context {
	if e.originContextP {
		return &e.originContext
	}

	return nil
}

func (e Error) OriginContextString() string {
	if e.originContextP {
		return fmt.Sprintf("%v", e.originContext)
	}

	return ""
}

func (e Error) LastMessage() string {
	return e.path[len(e.path)-1].Msg
}

func (e Error) Error() string {
	if e == nil || len(e.path) == 0 {
		return ""
	}

	msgs := make([]string, 0, len(e.path))

	for i := range e.path {
		msgs = append(msgs, e.path[i].Msg)
	}

	return strings.Join(msgs, "; ")
}

func (e Error) Class() ErrorClass {
	if e == nil {
		return NoError{}
	}

	if e.class == nil {
		return UnknownError{}
	}

	return e.class
}

func SetClass[T ErrorClass](errorToUpdate Error) {
	if errorToUpdate.class.Number() != 1 {
		return
	}

	var ec T
	errorToUpdate.class = ec
}

func (e Error) Is(target error) bool {
	if target == nil {
		return false
	}

	if tError, ok := target.(Error); ok { // nolint
		// Check for equivalent error classes first
		if e.class.Number() == tError.class.Number() {
			return true
		}
	}

	if e.originerror != nil {
		if errors.Is(target, e.originerror) { // nolint
			return true
		}
	}

	// Otherwise, check error strings.
	if e.Error() == target.Error() {
		return false
	}

	return false
}
