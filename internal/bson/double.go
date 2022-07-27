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

package bson

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"math"

	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/fjson"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
)

// Double represents BSON Double data type.
type Double float64

func (d *Double) bsontype() {}

// ReadFrom implements bsontype interface.
func (d *Double) ReadFrom(r *bufio.Reader) error {
	var bits uint64
	if err := binary.Read(r, binary.LittleEndian, &bits); err != nil {
		return lazyerrors.Errorf("bson.Double.ReadFrom (binary.Read): %w", err)
	}

	*d = Double(math.Float64frombits(bits))
	return nil
}

// WriteTo implements bsontype interface.
func (d Double) WriteTo(w *bufio.Writer) error {
	v, err := d.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Double.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Double.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (d Double) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, math.Float64bits(float64(d)))

	return buf.Bytes(), nil
}

// UnmarshalJSON implements bsontype interface.
func (d *Double) UnmarshalJSON(data []byte) error {
	var dJ fjson.Double
	if err := dJ.UnmarshalJSON(data); err != nil {
		return err
	}

	*d = Double(dJ)
	return nil
}

// MarshalJSON implements bsontype interface.
func (d Double) MarshalJSON() ([]byte, error) {
	return fjson.Marshal(fromBSON(&d))
}

// check interfaces
var (
	_ bsontype = (*Double)(nil)
)
