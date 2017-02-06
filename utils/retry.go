package utils

import "errors"

var errMaxRetriesReached = errors.New("exceeded retry limit")

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Retry keeps trying the function until the second argument
// returns false, or no error is returned.
func Retry(fn Func, maxRetries int) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if maxRetries > 0 && attempt > maxRetries {
			return errMaxRetriesReached
		}
	}
	return err
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}
