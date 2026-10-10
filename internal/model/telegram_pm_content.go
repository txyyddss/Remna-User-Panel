package model

import "errors"

// ErrPMContentExpired distinguishes expired new content from legacy reference-only jobs.
var ErrPMContentExpired = errors.New("pending Telegram PM content expired")

// ErrPMContentInvalid denotes an unreadable or unbound pending content envelope.
var ErrPMContentInvalid = errors.New("pending Telegram PM content invalid")
