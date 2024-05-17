// Copyright (C) 2022, 2023, 2024 by Blackcat Informatics® Inc.
//
// nolint: varnamelen,revive
package e

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ValueType int

const (
	ValTypeString ValueType = iota
)

type V struct {
	K string
	I interface{}
}

type Value interface{}
type Values []Value

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
//
// nolint: containedctx
type errorData struct {
	createdAt      time.Time
	originerror    error
	originContext  context.Context
	originContextP bool
	class          ErrorClass
	path           []PathElement
}

// The externally accessible error type.  Use this in your returns.
type Error = *errorData

// Resolve the set of values for this path element.
//
// nolint: nonamedreturns
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

	return
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

func Wrap(e Error, msg string) Error {
	if e == nil {
		return New[UnknownError](msg)
	}

	e.path = append(e.path, newPathElement(msg))

	return e
}

func WrapError[T ErrorClass](err error) Error {
	if err == nil {
		return New[NoError]("no error")
	}

	e := New[T](err.Error())
	e.originerror = err

	e.path[0].ValFunc = func() Values {
		return Values{
			"wrapped_error",
			V{K: "err", I: err},
		}
	}

	return e
}

func WrapErrorMsg[T ErrorClass](err error, msg string) Error {
	if err == nil {
		return New[NoError]("no error")
	}

	e := New[T](err.Error())
	e.path[0].ValFunc = func() Values {
		return Values{
			"wrapped_error",
			V{K: "err", I: err},
		}
	}
	e.path = append(e.path, newPathElement(msg))

	return e
}

func (e Error) AddValues(valFunc ValueFunc) Error {
	valFuncOrig := e.path[len(e.path)-1].ValFunc
	e.path[len(e.path)-1].ValFunc = func() Values {
		return append(valFuncOrig(), valFunc()...)
	}

	return e
}

func (e Error) AddValue(key string, val interface{}) Error {
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

func WrapWithVals[T ErrorClass](e Error, msg string, valFunc ValueFunc) Error {
	if e == nil {
		return New[T](msg)
	}

	e.path = append(e.path, newPathElement(msg))
	e.path[len(e.path)-1].ValFunc = valFunc

	return e
}

// FullWrap wraps an Error, adds a message and values and possibly updates missing context information.
func FullWrap[T ErrorClass](ctx context.Context, e Error, msg string, valFunc ValueFunc) Error {
	if !e.originContextP {
		e.originContext = ctx
		e.originContextP = true
	}

	return WrapWithVals[T](e, msg, valFunc)
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

// nolint: ireturn
func (e Error) Class() ErrorClass {
	if e == nil {
		return NoError{}
	}

	if e.class == nil {
		return UnknownError{}
	}

	return e.class
}

func SetClass[T ErrorClass](e Error) {
	if e.class.Number() != 1 {
		return
	}

	var ec T
	e.class = ec
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
