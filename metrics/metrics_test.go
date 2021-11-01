// Copyright 2013 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"testing"
)

func TestMetricToString(t *testing.T) {
	scenarios := []struct {
		name     string
		input    Metric
		expected string
	}{
		{
			name: "valid metric without __name__ label",
			input: Metric{
				"first_name":   "electro",
				"occupation":   "robot",
				"manufacturer": "westinghouse",
			},
			expected: `{first_name="electro", manufacturer="westinghouse", occupation="robot"}`,
		},
		{
			name: "valid metric with __name__ label",
			input: Metric{
				"__name__":     "electro",
				"occupation":   "robot",
				"manufacturer": "westinghouse",
			},
			expected: `electro{manufacturer="westinghouse", occupation="robot"}`,
		},
		{
			name: "empty metric with __name__ label",
			input: Metric{
				"__name__": "fooname",
			},
			expected: "fooname",
		},
		{
			name:     "empty metric",
			input:    Metric{},
			expected: "{}",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			actual := scenario.input.String()
			if actual != scenario.expected {
				t.Errorf("expected string output %s but got %s", actual, scenario.expected)
			}
		})
	}
}
