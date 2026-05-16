package core

import "errors"

var (
	ErrUnsupportedBank	= errors.New("unsupported bank statement")
	ErrEmptyPDF		= errors.New("empty pdf text")
	ErrInvalidFormat	= errors.New("invalid statement format")
)
