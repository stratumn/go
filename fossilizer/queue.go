// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package fossilizer

import (
	"context"
)

// FossilsQueue can be used by batch fossilizers to store data that's pending
// fossilization.
// StorageQueue implementations can use persistent storage (like cloud queues)
// to prevent loss of data in cloud micro-services architecture.
type FossilsQueue interface {
	// Push a fossil to the queue.
	Push(context.Context, *Fossil) error

	// Pop fossils from the queue.
	Pop(context.Context, int) ([]*Fossil, error)
}
