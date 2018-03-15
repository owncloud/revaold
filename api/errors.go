package api

import (
	"fmt"
)

type ErrorCode string

const (
	// StorageNotFoundErrorCode is used when a resource is not found.
	StorageNotFoundErrorCode ErrorCode = "STORAGE_ERROR_NOT_FOUND"

	// StorageAlreadyExistsErrorCode is used when a resource already exists.
	StorageAlreadyExistsErrorCode ErrorCode = "STORAGE_ERROR_ALREADY_EXISTS"

	// StoragePermissionDeniedErrorCode is used when it is not possible to acces the resource.
	StoragePermissionDeniedErrorCode ErrorCode = "STORAGE_ERROR_PERMISSION_DENIED"

	// ContextUserRequired requires an pkg.User object in the context
	ContextUserRequiredError ErrorCode = "CONTEXT_USER_REQUIRED"

	// PathInvalidError is used when a path is invalid, like not begging with /
	PathInvalidError ErrorCode = "PATH_INVALID_ERROR"

	// LinkNotFoundErrorCode is used when a resource is not found.
	PublicLinkNotFoundErrorCode ErrorCode = "PUBLIC_LINK_NOT_FOUND"

	PublicLinkInvalidExpireDateErrorCode ErrorCode = "PUBLIC_LINK_INVALID_EXPIRE_DATE"

	// StorageOperationNotSupported is used when some operation is not available on
	// the storage, like emptying the recycle bin
	StorageNotSupportedErrorCode ErrorCode = "STORAGE_NOT_SUPPORTED"

	UserNotFoundErrorCode ErrorCode = "USER_NOT_FOUND"
)

func NewError(code ErrorCode) Error {
	return Error{Code: code}
}

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e Error) WithMessage(msg string) Error {
	e.Message = msg
	return e
}

func (e Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	} else {
		return fmt.Sprintf("%s", e.Code)
	}
}
