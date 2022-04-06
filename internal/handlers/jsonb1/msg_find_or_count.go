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
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFindOrCount finds documents in a collection or view and returns a cursor to the selected documents
// or count the number of documents that matches the query filter.
func (h *storage) MsgFindOrCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	fmt.Println(document)

	var filter types.Document
	var sql, collection string

	var args []any

	m := document.Map()
	_, isFindOp := m["find"].(string)
	db := m["$db"].(string)

	fmt.Println(m)
	fmt.Println(m["projection"])

	var exclusion, projectBool bool

	if isFindOp { //enters here if find
		var projectionSQL string

		projectionIn, _ := m["projection"].(types.Document)
		//projectionIn.Set("ignoreKeys", true)
		fmt.Println("Projection")
		fmt.Println(projectionIn)
		projectionSQL, exclusion, projectBool, err = projection(projectionIn)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		//args = append(args, projectionArgs...)

		collection = m["find"].(string)
		filter, _ = m["filter"].(types.Document)
		//sql = fmt.Sprintf(`select %s FROM %s`, projectionSQL, pgx.Identifier{db, collection}.Sanitize())
		sql = fmt.Sprintf(`select %s FROM %s`, projectionSQL, collection)
	} else { // enters here if count
		collection = m["count"].(string)
		//filter, _ = m["query"].(types.Document)
		//sql = fmt.Sprintf(`select COUNT(*) FROM %s`, pgx.Identifier{db, collection}.Sanitize())
		sql = fmt.Sprintf(`select COUNT(*) FROM %s`, collection)
	}

	sort, _ := m["sort"].(types.Document)
	limit, _ := m["limit"].(int32)
	fmt.Println("Filter:")
	fmt.Println(filter)
	fmt.Println(filter.Keys())
	fmt.Println(filter)
	fmt.Println(filter.Map())
	for key := range filter.Map() {
		fmt.Println("key:")
		fmt.Println(key)
		fmt.Println(filter.Map()[key])
		sql += " WHERE "
		if strings.Contains(key, ".") {
			split := strings.Split(key, ".")
			count := 0
			for _, s := range split {
				if (len(split) - 1) == count {
					sql += "\"" + s + "\""
				} else {
					sql += "\"" + s + "\"."
				}
				count += 1
			}
		} else {
			sql += "\"" + key + "\""
		}

		sql += " = "
		//sql += placeholder.Next()
		value, _ := filter.Get(key)
		fmt.Println("value")
		fmt.Println(value)
		switch value := value.(type) {
		case string:
			args = append(args, value)
			sql += "'%s'"
		case int:
			fmt.Println("Here")
		case int64:
			fmt.Println("is Int")
			args = append(args, value)
		case int32:
			fmt.Println("int32")
			sql += "%d"
			//newValue, errorV := strconv.ParseInt(string(value), 10, 64)
			//if errorV != nil {
			//	fmt.Println("error")
			//}
			args = append(args, value)
		case types.Document:
			fmt.Println("is a document")
			fmt.Println(value)
			sql += "%s"
			argDoc, err := whereDocument(value)

			if err != nil {
				err = lazyerrors.Errorf("scalar: %w", err)
			}
			args = append(args, argDoc)
		case types.ObjectID:
			fmt.Println("is an Object")
			sql += "%s"
			var bOBJ []byte
			if bOBJ, err = bson.ObjectID(value).MarshalJSONHANA(); err != nil {
				err = lazyerrors.Errorf("scalar: %w", err)
			}
			fmt.Println("bObject")
			fmt.Println(bOBJ)
			//byt := make([]byte, hex.EncodedLen(len(value[:])))
			//fmt.Println("byt")
			//fmt.Println(byt)
			//fmt.Println(string(byt))
			//bstring := "{\"oid\": " + "'" + string(byt) + "'}"
			//fmt.Println("bstring")
			//fmt.Println(bstring)
			args = append(args, string(bOBJ))
		default:
			fmt.Println("Nothing")
		}
	}

	//no where so far
	//whereSQL, whereArgs, err := where(filter, &placeholder)
	//if err != nil {
	//	return nil, lazyerrors.Error(err)
	//}
	//args = append(args, whereArgs...)

	fmt.Println("args:")
	fmt.Println(args)

	//sql += whereSQL
	sqln := fmt.Sprintf(sql, args...)
	fmt.Println("sqln:")
	fmt.Println(sqln)

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
		// TODO https://github.com/FerretDB/FerretDB/issues/79
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgFind: negative limit values are not supported")
	}
	fmt.Println(sql)
	fmt.Println(sql, args)
	rows, err := h.hanaPool.QueryContext(ctx, fmt.Sprintf(sql, args...))
	//rows, err := h.hanaPool.QueryContext(ctx, sql, args...)
	if err != nil {
		fmt.Println("THE ERROR")
		return nil, lazyerrors.Error(err)
	}
	fmt.Println(rows)
	defer rows.Close()
	var reply wire.OpMsg
	if isFindOp { //nolint:nestif // FIXME: I have no idead to fix this lint
		var docs types.Array
		//docs := make([]types.Document, 0, 16)

		for {
			doc, err := nextRow(rows)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if doc == nil {
				break
			}

			//docs = append(docs, *doc)

			if err = docs.Append(*doc); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		fmt.Println("IS PROHECTBOOL True")
		if projectBool {
			fmt.Println("projectBool = true")
			err = projectDocuments(&docs, m["projection"].(types.Document), exclusion)
		}
		fmt.Println("DOCS")
		fmt.Println(docs)
		//firstBatch := types.MakeArray(len(docs))
		//for _, doc := range docs {
		//	if err = firstBatch.Append(doc); err != nil {
		//		fmt.Println("YES ERROR")
		//		return nil, err
		//	}
		//}
		//fmt.Println("firstBatch")
		//fmt.Println(firstBatch)

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
		// in psql, the SELECT * FROM table limit `x` ignores the value of the limit,
		// so, we need this `if` statement to support this kind of query `db.actor.find().limit(10).count()`
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
