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
	"encoding/json"
	"math"
	"testing"
)

func TestEqualValues(t *testing.T) {
	tests := map[string]struct {
		in1, in2 SampleValue
		want     bool
	}{
		"equal floats": {
			in1:  3.14,
			in2:  3.14,
			want: true,
		},
		"unequal floats": {
			in1:  3.14,
			in2:  3.1415,
			want: false,
		},
		"positive inifinities": {
			in1:  SampleValue(math.Inf(+1)),
			in2:  SampleValue(math.Inf(+1)),
			want: true,
		},
		"negative inifinities": {
			in1:  SampleValue(math.Inf(-1)),
			in2:  SampleValue(math.Inf(-1)),
			want: true,
		},
		"different inifinities": {
			in1:  SampleValue(math.Inf(+1)),
			in2:  SampleValue(math.Inf(-1)),
			want: false,
		},
		"number and infinity": {
			in1:  42,
			in2:  SampleValue(math.Inf(+1)),
			want: false,
		},
		"number and NaN": {
			in1:  42,
			in2:  SampleValue(math.NaN()),
			want: false,
		},
		"NaNs": {
			in1:  SampleValue(math.NaN()),
			in2:  SampleValue(math.NaN()),
			want: true, // !!!
		},
	}

	for name, test := range tests {
		got := test.in1.Equal(test.in2)
		if got != test.want {
			t.Errorf("Comparing %s, %f and %f: got %t, want %t", name, test.in1, test.in2, got, test.want)
		}
	}
}

func TestSamplePairJSON(t *testing.T) {
	input := []struct {
		plain string
		value SamplePair
	}{
		{
			plain: `[1234.567,"123.1"]`,
			value: SamplePair{
				Value:     123.1,
				Timestamp: 1234567,
			},
		},
	}

	for _, test := range input {
		b, err := json.Marshal(test.value)
		if err != nil {
			t.Error(err)
			continue
		}

		if string(b) != test.plain {
			t.Errorf("encoding error: expected %q, got %q", test.plain, b)
			continue
		}

		var sp SamplePair
		err = json.Unmarshal(b, &sp)
		if err != nil {
			t.Error(err)
			continue
		}

		if sp != test.value {
			t.Errorf("decoding error: expected %v, got %v", test.value, sp)
		}
	}
}
