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

package sql

import (
	"github.com/FerretDB/FerretDB/internal/hana"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	//"github.com/FerretDB/FerretDB/internal/pg"
)

//type storage struct {
//	pgPool *pg.Pool
//	l      *zap.SugaredLogger
//}

type storage struct {
	hanaPool *hana.Hpool
	l        *zap.SugaredLogger
}

func NewStorage(hanaPool *hana.Hpool, l *zap.SugaredLogger) common.Storage {
	return &storage{
		hanaPool: hanaPool,
		l:        l,
	}
}

//func NewStorage(pgPool *pg.Pool, l *zap.SugaredLogger) common.Storage {
//	return &storage{
//		pgPool: pgPool,
//		l:      l,
//	}
//}
