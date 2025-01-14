package errors

import "errors"

var ErrInvalidDuration = errors.New("time: invalid duration")
var ErrInvalidMapValue = errors.New("invalid map value")
var ErrNotAPointer = errors.New("not a pointer to a struct")
var ErrRequired = errors.New("required field not found")
var ErrNotEmpty = errors.New("environment variable is empty")
var ErrReadFile = errors.New("file read error")
var ErrLoadEnv = errors.New("error loading environment variables")
