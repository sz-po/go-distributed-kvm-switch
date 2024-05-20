package api

type Error struct {
	message  string
	httpCode int
}

func NewError(message string, httpCode int) Error {
	return Error{
		message:  message,
		httpCode: httpCode,
	}
}

func (e Error) Error() string {
	return e.message
}

func (e Error) HttpCode() int {
	return e.httpCode
}

var ErrObjectNotFound = NewError("object not found", 404)
var ErrObjectWithNameAlreadyExists = NewError("object with name already exists", 409)
var ErrDeletedObjectWithNameAlreadyExists = NewError("deleted object with name already exists", 409)
var ErrObjectAlreadyDeleted = NewError("object already deleted", 410)
var ErrObjectNotDeleted = NewError("object not deleted", 409)
