package tutorial

import "errors"

var (
	ErrCategoryRequired = errors.New("category is required")
	ErrCategoryTooLong  = errors.New("category must be 50 characters or less")
	ErrTitleRequired    = errors.New("title is required")
	ErrTitleTooLong     = errors.New("title must be 200 characters or less")
	ErrBodyRequired     = errors.New("body is required")
	ErrTutorialNotFound = errors.New("tutorial not found")
)
