// Copyright (C) 2022, 2023 by Blackcat Informatics® Inc.
//
// nolint:revive
package e

const defaultArea = "defaultErrors"

// NoError is for zero errors or nil errors.
type NoError struct{}

func (NoError) What() string   { return "NoError" }
func (NoError) Area() string   { return defaultArea }
func (NoError) Number() uint32 { return 0 } // nolint

// UnknownError is the default error class and indicates that no other
// class has been set.  This class should not be used for anything
// else or set by callers.
type UnknownError struct{}

func (UnknownError) What() string   { return "UnknownError" }
func (UnknownError) Area() string   { return defaultArea }
func (UnknownError) Number() uint32 { return 1 } // nolint

// LogicError is used in cases where internal or assumed logic has
// been violated.  Cases such as using a function incorrectly.
type LogicError struct{}

func (LogicError) What() string   { return "LogicError" }
func (LogicError) Area() string   { return defaultArea }
func (LogicError) Number() uint32 { return 2 } // nolint

// NotFoundError is used when what you are looking for is not there.
type NotFoundError struct{}

func (NotFoundError) What() string   { return "NotFoundError" }
func (NotFoundError) Area() string   { return defaultArea }
func (NotFoundError) Number() uint32 { return 3 } // nolint

// DataError is used for database errors or inconsistent data.
type DataError struct{}

func (DataError) What() string   { return "DataError" }
func (DataError) Area() string   { return defaultArea }
func (DataError) Number() uint32 { return 4 } // nolint

// PanicError is for wrapped panics.
type PanicError struct{}

func (PanicError) What() string   { return "PanicError" }
func (PanicError) Area() string   { return defaultArea }
func (PanicError) Number() uint32 { return 5 } // nolint

// FileError is for filesystem related errors.
type FileError struct{}

func (FileError) What() string   { return "FileError" }
func (FileError) Area() string   { return defaultArea }
func (FileError) Number() uint32 { return 6 } // nolint

// NetworkError is for network related errors.
type NetworkError struct{}

func (NetworkError) What() string   { return "NetworkError" }
func (NetworkError) Area() string   { return defaultArea }
func (NetworkError) Number() uint32 { return 7 } // nolint

// NetworkTempError is for transient network errors that can be recovered from later.
type NetworkTempError struct{}

func (NetworkTempError) What() string   { return "NetworkTempError" }
func (NetworkTempError) Area() string   { return defaultArea }
func (NetworkTempError) Number() uint32 { return 8 } // nolint

// ExecutionError is for when external execution fails for some reason.
type ExecutionError struct{}

func (ExecutionError) What() string   { return "ExecutionError" }
func (ExecutionError) Area() string   { return defaultArea }
func (ExecutionError) Number() uint32 { return 9 } // nolint

// APIError is for external API errors (not network errors).
type APIError struct{}

func (APIError) What() string   { return "APIError" }
func (APIError) Area() string   { return defaultArea }
func (APIError) Number() uint32 { return 10 } // nolint

// ValidationError is for when validation of data fails.
type ValidationError struct{}

func (ValidationError) What() string   { return "ValidationError" }
func (ValidationError) Area() string   { return defaultArea }
func (ValidationError) Number() uint32 { return 11 } // nolint

// StateError is for when we have a state violation or the application is in an incomplete state.
type StateError struct{}

func (StateError) What() string   { return "StateError" }
func (StateError) Area() string   { return defaultArea }
func (StateError) Number() uint32 { return 12 } // nolint
