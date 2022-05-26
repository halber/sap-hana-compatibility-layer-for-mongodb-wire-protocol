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

package jsonb1

import (
	"context"
	"fmt"

	"github.com/DocStore/HANA_HWY/internal/handlers/common"

	"github.com/DocStore/HANA_HWY/internal/types"
	"github.com/DocStore/HANA_HWY/internal/util/lazyerrors"
	"github.com/DocStore/HANA_HWY/internal/wire"
)

// MsgFindOrCount finds documents in a collection or view and returns a cursor to the selected documents
// or count the number of documents that matches the query filter.
func (h *storage) MsgFindOrCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"skip",
		"sort",
		"returnKey",
		"showRecordId",
		"tailable",
		"oplogReplay",
		"noCursorTimeout",
		"awaitData",
		"allowPartialResults",
		"collation",
		"allowDiskUse",
		"let",
		"hint",
		"batchSize",
		"singleBatch",
		"maxTimeMS",
		"readConcern",
		"max",
		"min",
		"comment",
	}

	if err := common.Unimplemented(&document, unimplementedFields...); err != nil {
		return nil, err
	}

	docMap := document.Map()

	if isPrintShardingStatus(docMap) {
		return nil, common.NewErrorMessage(common.ErrCommandNotFound, "no such command: printShardingStatus")
	}

	var filter types.Document
	var sql, collection string

	var args []any

	m := document.Map()
	_, isFindOp := m["find"].(string)
	db := m["$db"].(string)

	var exclusion, projectBool bool

	if isFindOp { //enters here if find
		var projectionSQL string

		projectionIn, _ := m["projection"].(types.Document)
		projectionSQL, exclusion, projectBool, err = common.Projection(projectionIn)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		collection = m["find"].(string)
		filter, _ = m["filter"].(types.Document)
		sql = fmt.Sprintf(`select %s FROM %s`, projectionSQL, collection)
	} else { // enters here if count
		collection = m["count"].(string)
		sql = fmt.Sprintf(`select COUNT(*) FROM %s`, collection)
	}

	sort, _ := m["sort"].(types.Document)
	limit, _ := m["limit"].(int32)

	whereSQL, whereArgs, err := common.WhereHANA(filter)
	if err != nil {
		return nil, err
	}

	sql += fmt.Sprintf(whereSQL, whereArgs...)

	sortMap := sort.Map()
	if len(sortMap) != 0 {
		sql += " ORDER BY"

		for i, k := range sort.Keys() {
			if i != 0 {
				sql += ","
			}

			sql += " \"%s\" "
			args = append(args, k)

			order := sortMap[k].(int32)
			if order > 0 {
				sql += " ASC"
			} else {
				sql += " DESC"
			}
		}
	}

	switch {
	case limit == 0:
		// undefined or zero - no limit
	case limit > 0:
		sql += " LIMIT %d"
		args = append(args, limit)
	default:
		// TODO https://github.com/DocStore/HANA_HWY/issues/79
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgFind: negative limit values are not supported")
	}
	rows, err := h.hanaPool.QueryContext(ctx, fmt.Sprintf(sql, args...))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()
	var reply wire.OpMsg
	if isFindOp { //nolint:nestif // FIXME: I have no idead to fix this lint
		var docs types.Array

		for {
			doc, err := nextRow(rows)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if doc == nil {
				break
			}

			if err = docs.Append(*doc); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		if projectBool {
			err = common.ProjectDocuments(&docs, m["projection"].(types.Document), exclusion)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		err = reply.SetSections(wire.OpMsgSection{
			Documents: []types.Document{types.MustMakeDocument(
				"cursor", types.MustMakeDocument(
					"firstBatch", &docs,
					"id", int64(0), // TODO
					"ns", db+"."+collection,
				),
				"ok", float64(1),
			)},
		})
	} else {
		var count int32
		for rows.Next() {
			err := rows.Scan(&count)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		if count > limit && limit != 0 {
			count = limit
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []types.Document{types.MustMakeDocument(
				"n", count,
				"ok", float64(1),
			)},
		})
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

func isPrintShardingStatus(docMap map[string]any) bool {

	if docMap["find"] == "shards" && docMap["$db"] == "config" {
		fmt.Println(1)
		return true
	} else if docMap["find"] == "mongos" && docMap["$db"] == "config" {
		fmt.Println(2)
		return true
	} else if docMap["find"] == "version" && docMap["$db"] == "config" {
		fmt.Println(3)
		return true
	}
	fmt.Println("NO")
	return false
}
