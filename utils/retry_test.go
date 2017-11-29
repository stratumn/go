// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
