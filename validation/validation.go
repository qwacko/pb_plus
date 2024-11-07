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
	Collection string `mapstructure:"collection"`
	Field      string `mapstructure:"field"`
	Filename   string `mapstructure:"filename"`
}

func ConfigureSchemaValidation(app *pocketbase.PocketBase, vAll *viper.Viper) {

	v := vAll.Sub("validation")

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)
	v.SetDefault("schema_dir", "./pb_schema")
	v.SetDefault("collection_name", "_schema")

	if !v.GetBool("enabled") {
		return
	}

	schemaDir := v.GetString("schema_dir")
	collectionName := v.GetString("collection_name")
	var schemaConfigs []SchemaConfig
	if err := v.UnmarshalKey("schema", &schemaConfigs); err != nil {
		log.Fatalf("Error unmarshalling schema configurations: %v", err)
	}
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {

		var collection *core.Collection

		if v.IsSet("view_rule") {
			viewRule := v.GetString("view_rule")

			collection = getOrCreateSchemaCollection(app, collectionName, &viewRule)
		} else {
			collection = getOrCreateSchemaCollection(app, collectionName, nil)
		}

		for _, config := range schemaConfigs {
			schemaPath := schemaDir + "/" + config.Filename
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				return fmt.Errorf("schema file does not exist: %s", schemaPath)
			}

			result, err := app.FindFirstRecordByFilter(collection, "table = {:table} && column = {:column}", dbx.Params{
				"table":  config.Collection,
				"column": config.Field,
			})

			if err != nil {
				if err == sql.ErrNoRows {
					new_record := core.NewRecord(collection)
					new_record.Set("table", config.Collection)
					new_record.Set("column", config.Field)

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
		if e.Collection.Name == collectionName {
			return apis.NewForbiddenError("You cannot update the schema table", "")
		}
		return e.Next()
	})

	app.OnCollectionUpdate(collectionName).BindFunc(func(e *core.CollectionEvent) error {
		// "e.HttpContext" is no longer available because "e" is the request event itself ...
		if v.IsSet("view_rule") {
			viewRule := v.GetString("view_rule")
			validateSchemaTableColumns(e.Collection, &viewRule)
		} else {
			validateSchemaTableColumns(e.Collection, nil)
		}

		return e.Next()
	})

	app.OnCollectionDeleteExecute(collectionName).BindFunc(func(e *core.CollectionEvent) error {
		return apis.NewForbiddenError("You cannot delete the schema table", "")
	})
	app.OnRecordUpdateRequest(collectionName).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot update records in the schema table", "")
	})
	app.OnRecordCreateRequest(collectionName).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot create a record in the schema table", "")
	})
	app.OnRecordDeleteRequest(collectionName).BindFunc(func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("You cannot delete a record in the schema table", "H")
	})

	// Add hooks for record creation and update to validate data
	app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {

		err := validateRecordData(app, e.Record, collectionName)
		if err != nil {
			return err
		}
		return e.Next()
	})

	app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
		err := validateRecordData(app, e.Record, collectionName)
		if err != nil {
			return err
		}
		return e.Next()
	})

	return
}
