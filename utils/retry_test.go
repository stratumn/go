package utils

import (
	"errors"
	"testing"
)

const retriesExpected = 10

func TestRetryWithError(t *testing.T) {
	retriesCount := 0

	err := Retry(func(attempt int) (bool, error) {
		retriesCount++
		return true, errors.New("error")
	}, retriesExpected)

	if !IsMaxRetries(err) {
		t.Errorf("Retry(): expected error to be Max Retries was %v", err)
	}

	if got, want := retriesCount, retriesExpected; got != want {
		t.Errorf("Retry(): expected %v retries, got %v", want, got)
	}
}

func TestRetryWithoutError(t *testing.T) {
	err := Retry(func(attempt int) (bool, error) {
		if attempt == retriesExpected-1 {
			return false, nil
		}
		return true, errors.New("error")
	}, retriesExpected)

	if err != nil {
		t.Errorf("Retry(): expected no error, got %v", err)
	}
}
