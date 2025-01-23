// Copyright 2025 walteh LLC
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

package status

// 🔄 ProgressIndicator defines how progress is displayed
type ProgressIndicator interface {
	Start(message string)
	Update(progress float64)
	Stop()
}

// 🔄 DefaultProgressIndicator implements ProgressIndicator with a simple spinner
type DefaultProgressIndicator struct {
	frames  []string
	current int
}

// NewDefaultProgressIndicator creates a new DefaultProgressIndicator
func NewDefaultProgressIndicator() *DefaultProgressIndicator {
	return &DefaultProgressIndicator{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (p *DefaultProgressIndicator) Start(message string) {
	// TODO: Implement spinner animation
}

func (p *DefaultProgressIndicator) Update(progress float64) {
	// TODO: Update spinner frame and progress percentage
}

func (p *DefaultProgressIndicator) Stop() {
	// TODO: Clear spinner line
}
