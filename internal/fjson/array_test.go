// Copyright 2021 FerretDB Inc.
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

package fjson

import (
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
)

func convertArray(a *types.Array) *Array {
	res := Array(*a)
	return &res
}

var arrayTestCases = []testCase{{
	name: "array_all",
	v: convertArray(types.MustNewArray(
		types.MustNewArray(),
		types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
		true,
		time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
		types.MustMakeDocument(),
		42.13,
		int32(42),
		int64(42),
		"foo",
		nil,
	)),
	j: `[[],{"$b":"Qg==","s":128},true,{"$d":1627378542123},{"$k":[]},{"$f":42.13},42,{"$l":"42"},"foo",null]`,
}}

func TestArray(t *testing.T) {
	t.Parallel()
	testJSON(t, arrayTestCases, func() fjsontype { return new(Array) })
}

func FuzzArrayJSON(f *testing.F) {
	fuzzJSON(f, arrayTestCases, func() fjsontype { return new(Array) })
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestCases, func() fjsontype { return new(Array) })
}
