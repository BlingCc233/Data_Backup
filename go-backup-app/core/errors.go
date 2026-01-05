package core

import "errors"

var ErrNoFilesSelected = errors.New("no files selected after applying filters")
var ErrNoChanges = errors.New("no changes detected since parent backup")
var ErrInvalidPassword = errors.New("invalid password")
