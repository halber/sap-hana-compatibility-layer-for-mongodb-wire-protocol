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

package handlers

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.wdf.sap.corp/DocStore/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/types"
	"github.wdf.sap.corp/DocStore/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
	"github.wdf.sap.corp/DocStore/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/wire"
)

// MsgHostInfo returns an OpMsg with the host information.
func (h *Handler) MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	now := time.Now().UTC()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"system", types.MustMakeDocument(
				"currentTime", now,
				"hostname", hostname,
				"cpuAddrSize", int32(strconv.IntSize),
				"numCores", int32(runtime.NumCPU()),
				"cpuArch", runtime.GOARCH,
				"numaEnabled", false,
			),
			"os", types.MustMakeDocument(
				"type", cases.Title(language.English).String(runtime.GOOS),
			),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
