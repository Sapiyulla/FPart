package user

import "fpart/internal/pkg/errs"

var ErrUserNotFound error = &errs.ErrNotFound{Resource: "user"}
