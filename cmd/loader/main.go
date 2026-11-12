/*
 * Script to dump GORM model schema as SQL statements
 */
package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

//go:embed schema.sql
var enums string

func main() {
	stmts, err := gormschema.New("postgres", gormschema.WithConfig(database.GormConfig)).
		Load(models.AllModels...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}

	// Schema, then enum types (unqualified, like the GORM columns that use them).
	fmt.Fprintf(os.Stdout, "CREATE SCHEMA IF NOT EXISTS %[1]s;\nSET search_path TO %[1]s;\n%s\n", database.SchemaName, enums)
	io.WriteString(os.Stdout, stmts)
}
