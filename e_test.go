// Copyright (C) 2023-2023 by Blackcat Informatics® Inc.
// nolint
package e_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/paudley/e"
	. "github.com/smartystreets/goconvey/convey"
)

func cBad() string {
	s, _ := e.CallLocation()
	return s
}

var cBadS = cBad()

func fGood() e.Error {
	return nil
}

func fBad() e.Error {
	return e.New[e.UnknownError]("oops")
}

var (
	goErr1 = errors.New("goErr1")
	goErr2 = errors.New("goErr2")
)

func fBadNested() e.Error {
	if err := fBad(); err != nil {
		return e.Wrap(err, "second level oops")
	}

	return nil
}

func fBadNested2() e.Error {
	return e.Wrap(nil, "second level oops, no error")
}

func passThru(err e.Error) error {
	return err
}

func fBadValues() e.Error {
	return e.NewWithVals[e.NotFoundError]("moo", func() e.Values {
		return e.Values{
			"foo",
			e.V{K: "bar", I: 1929394},
			e.V{"baz", "for"},
		}
	})
}

func fBadNestedValues() e.Error {
	if err := fBadNested(); err != nil {
		return e.WrapWithVals[e.DataError](err, "mid level error string", func() e.Values { return e.Values{"ducks"} })
	}

	return nil
}

func fBadNested3() e.Error {
	if err := fBadNestedValues(); err != nil {
		return e.Wrap(err, "top level error string")
	}

	return nil
}

func fPanicValFunc() e.Error {
	return e.NewWithVals[e.LogicError]("moo", func() e.Values {
		panic("dunce")
	})
}

func fPanicValFunc2() e.Error {
	return e.NewWithVals[e.LogicError]("moo", func() e.Values {
		panic(fmt.Errorf("dunce2"))
	})
}

func fNilValFunc() e.Error {
	return e.NewWithVals[e.NotFoundError]("moo", nil)
}

func fContextError() e.Error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cval", "stringvalue1")

	return e.NewWithContext[e.NotFoundError](ctx, "cerr")
}

func fFull() e.Error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cval2", "stringvalue2")

	return e.Full[e.LogicError](ctx, "error msg goes here", func() e.Values {
		return e.Values{
			"meeses",
		}
	})
}

const defaultArea = "defaultErrors"

func TestErrorCreation(t *testing.T) {
	t.Parallel()
	Convey("Verify that error creation works.", t, func() {
		Convey("check the good function returns work as expected", func() {
			err := fGood()
			So(err, ShouldBeNil)
			So(err.Error(), ShouldEqual, "")
			So(err.Class().Number(), ShouldEqual, 0)
			So(err.Class().What(), ShouldEqual, "NoError")
			So(err.Class().Area(), ShouldEqual, defaultArea)
		})
		Convey("check the bad function returns work as expected", func() {
			err := fBad()
			So(err, ShouldNotBeNil)
			path := err.Path()
			// Single path element.
			So(len(path), ShouldEqual, 1)
			So(path[0].FuncName, ShouldEqual, "github.com/paudley/e_test.fBad")
			So(path[0].Msg, ShouldEqual, "oops")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
			So(err.Is(fBad()), ShouldBeTrue)
			So(err.Class(), ShouldNotBeNil)
			cl := err.Class()
			So(cl.What(), ShouldEqual, "UnknownError")
			So(cl.Area(), ShouldEqual, "defaultErrors")
			So(cl.Number(), ShouldEqual, 1)
			So(err.Is(err), ShouldBeTrue)
			So(err.Is(goErr1), ShouldBeFalse)
			So(err.Is(goErr2), ShouldBeFalse)
		})
		Convey("check wrapping interaction with std errors", func() {
			err := e.WrapError[e.LogicError](goErr1)
			So(err.Error(), ShouldEqual, "goErr1")
			So(err.Is(goErr1), ShouldBeTrue)
			So(err.Is(goErr2), ShouldBeFalse)
			So(err.Class(), ShouldEqual, e.LogicError{})
		})
		Convey("check second level simple wrapping with error", func() {
			err := fBadNested()
			So(err, ShouldNotBeNil)
			path := err.Path()
			So(len(path), ShouldEqual, 2)
			So(path[0].FuncName, ShouldEqual, "github.com/paudley/e_test.fBad")
			So(path[0].Msg, ShouldEqual, "oops")
			So(path[1].FuncName, ShouldEqual, "github.com/paudley/e_test.fBadNested")
			So(path[1].Msg, ShouldEqual, "second level oops")
			So(err.Error(), ShouldEqual, "oops; second level oops")
			// Check it survives error conversion
			errP := passThru(err)
			So(errP.Error(), ShouldEqual, "oops; second level oops")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
			So(err.Is(goErr1), ShouldBeFalse)
			So(err.Is(goErr2), ShouldBeFalse)
		})
		Convey("check second level simple wrapping with nil error", func() {
			err := fBadNested2()
			So(err, ShouldNotBeNil)
			path := err.Path()
			So(len(path), ShouldEqual, 1)
			So(path[0].FuncName, ShouldEqual, "github.com/paudley/e_test.fBadNested2")
			So(path[0].Msg, ShouldEqual, "second level oops, no error")
			So(err.Error(), ShouldEqual, "second level oops, no error")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
		})
		Convey("check valfunc mechanics", func() {
			err := fBadValues()
			So(err, ShouldNotBeNil)
			for _, path := range err.Path() {
				vals := path.Values()
				So(len(vals), ShouldEqual, 3)
				So(vals[0], ShouldEqual, "foo")
				v1, convOK := vals[1].(e.V)
				So(convOK, ShouldBeTrue)
				So(v1.K, ShouldEqual, "bar")
				So(v1.I, ShouldEqual, 1929394)
				v2, convOK := vals[2].(e.V)
				So(convOK, ShouldBeTrue)
				So(v2.K, ShouldEqual, "baz")
				So(v2.I, ShouldEqual, "for")
				// Check for empty context.
				So(err.OriginContext(), ShouldBeNil)
			}
		})
		Convey("check upper level value addition", func() {
			err := fBadNested3()
			So(err, ShouldNotBeNil)
			path := err.Path()
			So(len(path), ShouldEqual, 4)
			So(path[0].FuncName, ShouldEqual, "github.com/paudley/e_test.fBad")
			So(path[0].Msg, ShouldEqual, "oops")
			So(path[1].FuncName, ShouldEqual, "github.com/paudley/e_test.fBadNested")
			So(path[1].Msg, ShouldEqual, "second level oops")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
			vals := path[2].Values()
			So(len(vals), ShouldEqual, 1)
			So(vals[0], ShouldEqual, "ducks")
		})
		Convey("make sure that valfunc swallows panics (string)", func() {
			So(func() {
				err := fPanicValFunc()
				So(err, ShouldNotBeNil)
				_ = err.Path()[0].Values()
			}, ShouldNotPanic)
			err := fPanicValFunc()
			So(err, ShouldNotBeNil)
			vals := err.Path()[0].Values()
			So(len(vals), ShouldEqual, 1)
			v, convOK := vals[0].(e.V)
			So(convOK, ShouldBeTrue)
			So(v.K, ShouldEqual, "panic")
			So(v.I, ShouldEqual, "PANIC in ValFunc: dunce")
			So(err.LastMessage(), ShouldEqual, "moo")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
		})
		Convey("make sure that valfunc swallows panics (error)", func() {
			So(func() {
				err := fPanicValFunc2()
				So(err, ShouldNotBeNil)
				_ = err.Path()[0].Values()
			}, ShouldNotPanic)
			err := fPanicValFunc2()
			So(err, ShouldNotBeNil)
			vals := err.Path()[0].Values()
			So(len(vals), ShouldEqual, 1)
			v, convOK := vals[0].(e.V)
			So(convOK, ShouldBeTrue)
			So(v.K, ShouldEqual, "panic")
			So(v.I, ShouldEqual, "PANIC in ValFunc: dunce2")
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
		})
		Convey("make sure that we tolergage a nil val func", func() {
			err := fNilValFunc()
			So(err, ShouldNotBeNil)
			vals := err.Path()[0].Values()
			So(len(vals), ShouldEqual, 0)
			// Check for empty context.
			So(err.OriginContext(), ShouldBeNil)
		})
		Convey("make sure that context is kept if provided", func() {
			err := fContextError()
			So(err, ShouldNotBeNil)
			octx := err.OriginContext()
			So(octx, ShouldNotBeNil)
			So(fmt.Sprintf("%v", *octx), ShouldEqual, "context.Background.WithValue(type string, val stringvalue1)")
			So(err.OriginContextString(), ShouldEqual, "context.Background.WithValue(type string, val stringvalue1)")
			err2 := fBad()
			So(err2.OriginContextString(), ShouldEqual, "")
		})
		Convey("check full error generation", func() {
			err := fFull()
			So(err, ShouldNotBeNil)
			octx := err.OriginContext()
			So(octx, ShouldNotBeNil)
			So(fmt.Sprintf("%v", *octx), ShouldEqual, "context.Background.WithValue(type string, val stringvalue2)")
			path := err.Path()
			So(len(path), ShouldEqual, 1)
			So(path[0].FuncName, ShouldEqual, "github.com/paudley/e_test.fFull")
			vals := path[0].Values()
			So(len(vals), ShouldEqual, 1)
			So(vals[0], ShouldEqual, "meeses")
		})
		Convey("check wrapped error messages", func() {
			wraperr := fmt.Errorf("moose")
			So(wraperr, ShouldNotBeNil)
			err := e.WrapError[e.LogicError](wraperr)
			path := err.Path()
			So(path[0].Msg, ShouldEqual, "moose")
			err2 := e.WrapError[e.LogicError](nil)
			path2 := err2.Path()
			So(path2[0].Msg, ShouldEqual, "no error")
		})
	})
}
