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

	"github.wdf.sap.corp/DocStore/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/fjson"
	"github.wdf.sap.corp/DocStore/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
)

// Int64 represents BSON Int64 data type.
type Int64 int64

func (i *Int64) bsontype() {}

// ReadFrom implements bsontype interface.
func (i *Int64) ReadFrom(r *bufio.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, i); err != nil {
		return lazyerrors.Errorf("bson.Int64.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (i Int64) WriteTo(w *bufio.Writer) error {
	v, err := i.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Int64.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Int64.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (i Int64) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, i)

	return buf.Bytes(), nil
}

// UnmarshalJSON implements bsontype interface.
func (i *Int64) UnmarshalJSON(data []byte) error {
	var iJ fjson.Int64
	if err := iJ.UnmarshalJSON(data); err != nil {
		return err
	}

	*i = Int64(iJ)
	return nil
}

// MarshalJSON implements bsontype interface.
func (i Int64) MarshalJSON() ([]byte, error) {
	return fjson.Marshal(fromBSON(&i))
}

// check interfaces
var (
	_ bsontype = (*Int64)(nil)
)
