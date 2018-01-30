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

package validator

import (
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
)

var (
	// ErrMissingProcess is returned when the process name is missing for schema validation.
	ErrMissingProcess = errors.New("schema validation requires a process")

	// ErrMissingLinkType is returned when the link type is missing for schema validation.
	ErrMissingLinkType = errors.New("schema validation requires a link type")
)

type validatorBaseConfig struct {
	Process  string
	LinkType string
}

func newValidatorBaseConfig(process, linkType string) (*validatorBaseConfig, error) {
	if len(process) == 0 {
		return nil, ErrMissingProcess
	}

	if len(linkType) == 0 {
		return nil, ErrMissingLinkType
	}
	return &validatorBaseConfig{Process: process, LinkType: linkType}, nil
}

// shouldValidate returns true if the link matches the validator's process
// and type. Otherwise the link is considered valid because this validator
// doesn't apply to it.
func (bv *validatorBaseConfig) shouldValidate(link *cs.Link) bool {
	linkProcess, ok := link.Meta["process"].(string)
	if !ok {
		log.Debug("No process found in link %v", link)
		return false
	}

	if linkProcess != bv.Process {
		return false
	}

	linkAction, ok := link.Meta["action"].(string)
	if !ok {
		log.Debug("No action found in link %v", link)
		return false
	}

	if linkAction != bv.LinkType {
		return false
	}

	return true
}
