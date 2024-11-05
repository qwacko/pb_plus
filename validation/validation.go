package validation

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type SchemaConfig struct {
	Table    string `mapstructure:"table"`
	Column   string `mapstructure:"column"`
	Filename string `mapstructure:"filename"`
}

func ConfigureSchemaValidation(app *pocketbase.PocketBase, vAll *viper.Viper) {

	v := vAll.Sub("validation")

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)
	v.SetDefault("schemaDir", "./pb_schema")
	v.SetDefault("table", "_schema")
	v.SetDefault("viewAuthOnly", true)

	if !v.GetBool("enabled") {
		return
	}

	schemaDir := v.GetString("schemaDir")
	schemaTable := v.GetString("table")
	var schemaConfigs []SchemaConfig
	if err := v.UnmarshalKey("schema", &schemaConfigs); err != nil {
		log.Fatalf("Error unmarshalling schema configurations: %v", err)
	}
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {

		var collection *core.Collection

		if v.IsSet("viewRule") {
			viewRule := v.GetString("viewRule")

			collection = getOrCreateSchemaCollection(app, schemaTable, &viewRule)
		} else {
			collection = getOrCreateSchemaCollection(app, schemaTable, nil)
		}

		for _, config := range schemaConfigs {
			schemaPath := schemaDir + "/" + config.Filename
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				return fmt.Errorf("schema file does not exist: %s", schemaPath)
			}

			result, err := app.FindFirstRecordByFilter(collection, "table = {:table} && column = {:column}", dbx.Params{
				"table":  config.Table,
				"column": config.Column,
			})

			if err != nil {
				if err == sql.ErrNoRows {
					new_record := core.NewRecord(collection)
					new_record.Set("table", config.Table)
					new_record.Set("column", config.Column)

					schemaContent, schemaHash, err := readAndValidateSchema(schemaPath)
					if err != nil {
						return err
					}

					new_record.Set("hash", schemaHash)
					new_record.Set("schema", schemaContent)

					if err = app.Save(new_record); err != nil {
						return err
					}
					continue
				} else {
					return err
				}
			} else {
				// Update the schema
				schemaContent, schemaHash, err := readAndValidateSchema(schemaPath)
				if err != nil {
					return err
				}

				currentHash := result.GetString("hash")
				if currentHash != schemaHash {
					log.Println("Updating schema")
					result.Set("hash", schemaHash)
					result.Set("schema", schemaContent)
					if err = app.Save(result); err != nil {
						log.Printf("Error saving schema: %v", err)
						return err
					}
				}
			}
		}

		return e.Next()
	})

	app.OnCollectionUpdateRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
		if e.Collection.Name == schemaTable {
			return apis.NewForbiddenError("You cannot update the schema table", "")
		}
		return e.Next()
	})

	app.OnCollectionUpdate(schemaTable).BindFunc(func(e *core.CollectionEvent) error {
		// "e.HttpContext" is no longer available because "e" is the request event itself ...
		if v.IsSet("viewRule") {
			viewRule := v.GetString("viewRule")
			validateSchemaTableColumns(e.Collection, &viewRule)
		} else {
			validateSchemaTableColumns(e.Collection, nil)
		}

		return e.Next()
	})

	app.OnCollectionDeleteExecute(schemaTable).BindFunc(func(e *core.CollectionEvent) error {
		return apis.NewForbiddenError("You cannot delete the schema table", "")
	})
	app.OnRecordUpdateRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot update records in the schema table", "")
	})
	app.OnRecordCreateRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot create a record in the schema table", "")
	})
	app.OnRecordDeleteRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot delete a record in the schema table", "H")
	})

	// Add hooks for record creation and update to validate data
	app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {

		err := validateRecordData(app, e.Record, schemaTable)
		if err != nil {
			return err
		}
		return e.Next()
	})

	app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
		err := validateRecordData(app, e.Record, schemaTable)
		if err != nil {
			return err
		}
		return e.Next()
	})

	return
}
