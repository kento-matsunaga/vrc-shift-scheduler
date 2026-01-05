package announcement

import "errors"

var (
	ErrTitleRequired    = errors.New("title is required")
	ErrTitleTooLong     = errors.New("title must be 200 characters or less")
	ErrBodyRequired     = errors.New("body is required")
	ErrAnnouncementNotFound = errors.New("announcement not found")
)
