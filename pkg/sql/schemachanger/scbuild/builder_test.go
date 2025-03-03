// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package scbuild_test

import (
	"bufio"
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/datadriven"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestBuilderAlterTable(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx := context.Background()

	datadriven.Walk(t, filepath.Join("testdata"), func(t *testing.T, path string) {
		for _, depsType := range []struct {
			name                string
			dependenciesWrapper func(*testing.T, serverutils.TestServerInterface, *sqlutils.SQLRunner, func(scbuild.Dependencies))
		}{
			{
				"sql_dependencies",
				func(t *testing.T, s serverutils.TestServerInterface, tdb *sqlutils.SQLRunner, fn func(scbuild.Dependencies)) {
					sctestutils.WithBuilderDependenciesFromTestServer(s, fn)
				},
			},
			{
				"test_dependencies",
				func(t *testing.T, s serverutils.TestServerInterface, tdb *sqlutils.SQLRunner, fn func(scbuild.Dependencies)) {
					fn(sctestdeps.NewTestDependencies(ctx, t, tdb, nil /* testingKnobs */, nil /* statements */))
				},
			},
		} {
			t.Run(depsType.name, func(t *testing.T) {
				s, sqlDB, _ := serverutils.StartServer(t, base.TestServerArgs{})
				defer s.Stopper().Stop(ctx)

				tdb := sqlutils.MakeSQLRunner(sqlDB)

				datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
					return run(ctx, t, d, s, tdb, depsType.dependenciesWrapper)
				})
			})
		}
	})
}

func run(
	ctx context.Context,
	t *testing.T,
	d *datadriven.TestData,
	s serverutils.TestServerInterface,
	tdb *sqlutils.SQLRunner,
	withDependencies func(*testing.T, serverutils.TestServerInterface, *sqlutils.SQLRunner, func(scbuild.Dependencies)),
) string {
	switch d.Cmd {
	case "create-table", "create-view", "create-type", "create-sequence", "create-schema", "create-database":
		stmts, err := parser.Parse(d.Input)
		require.NoError(t, err)
		require.Len(t, stmts, 1)
		tableName := ""
		switch node := stmts[0].AST.(type) {
		case *tree.CreateTable:
			tableName = node.Table.String()
		case *tree.CreateSequence:
			tableName = node.Name.String()
		case *tree.CreateView:
			tableName = node.Name.String()
		case *tree.CreateType:
			tableName = ""
		case *tree.CreateSchema:
			tableName = ""
		case *tree.CreateDatabase:
			tableName = ""
		default:
			t.Fatal("not a supported CREATE statement")
		}
		tdb.Exec(t, d.Input)

		if len(tableName) > 0 {
			var tableID descpb.ID
			tdb.QueryRow(t, fmt.Sprintf(`SELECT '%s'::REGCLASS::INT`, tableName)).Scan(&tableID)
			if tableID == 0 {
				t.Fatalf("failed to read ID of new table %s", tableName)
			}
			t.Logf("created relation with id %d", tableID)
		}

		return ""
	case "build":
		var outputNodes scpb.State
		withDependencies(t, s, tdb, func(deps scbuild.Dependencies) {
			stmts, err := parser.Parse(d.Input)
			require.NoError(t, err)
			for i := range stmts {
				outputNodes, err = scbuild.Build(ctx, deps, outputNodes, stmts[i].AST)
				require.NoError(t, err)
			}
		})
		return marshalNodes(t, outputNodes)

	case "unimplemented":
		withDependencies(t, s, tdb, func(deps scbuild.Dependencies) {
			stmts, err := parser.Parse(d.Input)
			require.NoError(t, err)
			require.Len(t, stmts, 1)

			stmt := stmts[0]
			alter, ok := stmt.AST.(*tree.AlterTable)
			require.Truef(t, ok, "not an ALTER TABLE statement: %s", stmt.SQL)

			_, err = scbuild.Build(ctx, deps, nil, alter)
			require.Truef(t, scbuild.HasNotImplemented(err), "expected unimplemented, got %v", err)
		})
		return ""

	default:
		return fmt.Sprintf("unknown command: %s", d.Cmd)
	}
}

// indentText indents text for formatting out marshaled data.
func indentText(input string, tab string) string {
	result := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		result.WriteString(tab)
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

// marshalNodes marshals a scpb.State to YAML.
func marshalNodes(t *testing.T, nodes scpb.State) string {
	var sortedEntries []string
	for _, node := range nodes {
		var buf bytes.Buffer
		require.NoError(t, (&jsonpb.Marshaler{}).Marshal(&buf, node.Target.Element()))
		target := make(map[string]interface{})
		require.NoError(t, gojson.Unmarshal(buf.Bytes(), &target))
		entry := strings.Builder{}
		entry.WriteString("- ")
		entry.WriteString(node.Target.Direction.String())
		entry.WriteString(" ")
		entry.WriteString(screl.ElementString(node.Element()))
		entry.WriteString("\n")
		entry.WriteString(indentText(fmt.Sprintf("state: %s\n", node.Status.String()), "  "))
		entry.WriteString(indentText("details:\n", "  "))
		out, err := yaml.Marshal(target)
		require.NoError(t, err)
		entry.WriteString(indentText(string(out), "    "))
		sortedEntries = append(sortedEntries, entry.String())
	}
	// Sort the output buffer of nodes for determinism.
	result := strings.Builder{}
	sort.Strings(sortedEntries)
	for _, entry := range sortedEntries {
		result.WriteString(entry)
	}
	return result.String()
}
