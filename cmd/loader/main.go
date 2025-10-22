/*
 * Script to dump GORM model schema as SQL statements
 */
package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

func main() {
	stmts, err := gormschema.New("postgres", gormschema.WithConfig(database.GormConfig)).
		Load(
			&models.ContestParticipantGroup{},
			&models.ContestParticipant{},
			&models.ContestPendingRegistration{},
			&models.ContestTask{},
			&models.Contest{},
			&models.File{},
			&models.Group{},
			&models.LanguageConfig{},
			&models.QueueMessage{},
			&models.SubmissionResult{},
			&models.Submission{},
			&models.TaskGroup{},
			&models.TaskUser{},
			&models.Task{},
			&models.TestCase{},
			&models.TestResult{},
			&models.UserGroup{},
			&models.User{},
		)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}

	// Prepend schema creation to the output
	schemaCreation := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;\n", database.SchemaName)
	io.WriteString(os.Stdout, schemaCreation)
	io.WriteString(os.Stdout, stmts)
}
