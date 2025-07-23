package utils

import "errors"

var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrUnknownJSONFields = errors.New("unknown json fields")
