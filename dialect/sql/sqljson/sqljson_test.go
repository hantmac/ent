// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package sqljson_test

import (
	"strconv"
	"testing"

	"github.com/facebook/ent/dialect"
	"github.com/facebook/ent/dialect/sql"
	"github.com/facebook/ent/dialect/sql/sqljson"
	"github.com/stretchr/testify/require"
)

func TestWritePath(t *testing.T) {
	tests := []struct {
		input     sql.Querier
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			input: sql.Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", 1, sqljson.Path("b", "c", "[1]", "d"), sqljson.Cast("int"))),
			wantQuery: `SELECT * FROM "users" WHERE CAST("a"->'b'->'c'->1->'d' AS int) = $1`,
			wantArgs:  []interface{}{1},
		},
		{
			input: sql.Dialect(dialect.MySQL).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", "a", sqljson.DotPath("b.c[1].d"))),
			wantQuery: "SELECT * FROM `users` WHERE JSON_EXTRACT(`a`, \"$.b.c[1].d\") = ?",
			wantArgs:  []interface{}{"a"},
		},
		{
			input: sql.Dialect(dialect.MySQL).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", "a", sqljson.DotPath("b.\"c[1]\".d[1][2].e"))),
			wantQuery: "SELECT * FROM `users` WHERE JSON_EXTRACT(`a`, \"$.b.\"c[1]\".d[1][2].e\") = ?",
			wantArgs:  []interface{}{"a"},
		},
		{
			input: sql.Select("*").
				From(sql.Table("test")).
				Where(sqljson.HasKey("j", sqljson.DotPath("a.*.c"))),
			wantQuery: "SELECT * FROM `test` WHERE JSON_EXTRACT(`j`, \"$.a.*.c\") IS NOT NULL",
		},
		{
			input: sql.Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("test")).
				Where(sqljson.HasKey("j", sqljson.DotPath("a.b.c"))),
			wantQuery: `SELECT * FROM "test" WHERE "j"->'a'->'b'->'c' IS NOT NULL`,
		},
		{
			input: sql.Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("test")).
				Where(sql.And(
					sql.EQ("e", 10),
					sqljson.ValueEQ("a", 1, sqljson.DotPath("b.c")),
				)),
			wantQuery: `SELECT * FROM "test" WHERE "e" = $1 AND "a"->'b'->'c' = $2`,
			wantArgs:  []interface{}{10, 1},
		},
		{
			input: sql.Dialect(dialect.MySQL).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", "a", sqljson.Path("b", "c", "[1]", "d"), sqljson.Unquote(true))),
			wantQuery: "SELECT * FROM `users` WHERE JSON_UNQUOTE(JSON_EXTRACT(`a`, \"$.b.c[1].d\")) = ?",
			wantArgs:  []interface{}{"a"},
		},
		{
			input: sql.Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", "a", sqljson.Path("b", "c", "[1]", "d"), sqljson.Unquote(true))),
			wantQuery: `SELECT * FROM "users" WHERE "a"->'b'->'c'->1->>'d' = $1`,
			wantArgs:  []interface{}{"a"},
		},
		{
			input: sql.Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users")).
				Where(sqljson.ValueEQ("a", 1, sqljson.Path("b", "c", "[1]", "d"), sqljson.Cast("int"))),
			wantQuery: `SELECT * FROM "users" WHERE CAST("a"->'b'->'c'->1->'d' AS int) = $1`,
			wantArgs:  []interface{}{1},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			query, args := tt.input.Query()
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		input    string
		wantPath []string
		wantErr  bool
	}{
		{
			input:    "a.b.c",
			wantPath: []string{"a", "b", "c"},
		},
		{
			input:    "a[1][2]",
			wantPath: []string{"a", "[1]", "[2]"},
		},
		{
			input:    "a[1][2].b",
			wantPath: []string{"a", "[1]", "[2]", "b"},
		},
		{
			input:    `a."b.c[0]"`,
			wantPath: []string{"a", `"b.c[0]"`},
		},
		{
			input:    `a."b.c[0]".d`,
			wantPath: []string{"a", `"b.c[0]"`, "d"},
		},
		{
			input: `...`,
		},
		{
			input:    `.a.b.`,
			wantPath: []string{"a", "b"},
		},
		{
			input:   `a."`,
			wantErr: true,
		},
		{
			input:   `a[`,
			wantErr: true,
		},
		{
			input:   `a[a]`,
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			path, err := sqljson.ParsePath(tt.input)
			require.Equal(t, tt.wantPath, path)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
