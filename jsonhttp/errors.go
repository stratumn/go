package jsonhttp

import (
	"encoding/json"
)

var (
	// ErrInternalServer is an error for when an internal server occurs.
	ErrInternalServer = ErrHTTP{"internal server error", 500}
	// ErrBadRequest is an error for when a bad request occurs.
	ErrBadRequest = ErrHTTP{"bad request", 400}
	// ErrUnauthorized is an error for when an unauthorized request occurs.
	ErrUnauthorized = ErrHTTP{"unauthorized", 401}
	// ErrNotFound is an error for when something isn't found.
	ErrNotFound = ErrHTTP{"not found", 404}
)

// ErrHTTP is an error with an HTTP status.
type ErrHTTP struct {
	Msg    string `json:"error"`
	Status int    `json:"status"`
}

// Error implements an error.
func (e *ErrHTTP) Error() string {
	return e.Msg
}

// JSONEncode marshals an error to JSON.
func (e *ErrHTTP) JSONEncode() []byte {
	js, err := json.Marshal(e)

	if err != nil {
		msg := `{"error:": "an internal server error occured", "status": 500}`
		return []byte(msg)
	}

	return js
}
