package validation

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
)

type SchemaConfig struct {
	Table    string `mapstructure:"table"`
	Column   string `mapstructure:"column"`
	Filename string `mapstructure:"filename"`
}

// Refactored hashMD5 function for efficiency
func hashMD5(input string) string {
	hasher := md5.Sum([]byte(input))
	return hex.EncodeToString(hasher[:])
}

// Helper function to read and validate schema files
func readAndValidateSchema(schemaPath string) (string, string, error) {
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Printf("Error reading schema file %s: %v", schemaPath, err)
		return "", "", err
	}

	if err := validateJSONSchema(string(schemaContent)); err != nil {
		log.Printf("Invalid JSON schema in %s: %v", schemaPath, err)
		return "", "", err
	}

	schemaHash := hashMD5(string(schemaContent))
	return string(schemaContent), schemaHash, nil
}

func validateJSONSchema(schemaContent string) error {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaContent), &schema); err != nil {
		return errors.New("invalid JSON schema")
	}

	loader := gojsonschema.NewStringLoader(schemaContent)
	_, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return errors.New("invalid JSON schema")
	}
	return nil
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
			return errors.New("You cannot update the schema table")
		}
		return e.Next()
	})

	app.OnCollectionUpdate(schemaTable).BindFunc(func(e *core.CollectionEvent) error {
		// "e.HttpContext" is no longer available because "e" is the request event itself ...
		if v.IsSet("viewRule") {
			viewRule := v.GetString("viewRule")
			validateSchemaTableColumns(app, e.Collection, &viewRule)
		} else {
			validateSchemaTableColumns(app, e.Collection, nil)
		}

		return e.Next()
	})

	app.OnCollectionDeleteExecute(schemaTable).BindFunc(func(e *core.CollectionEvent) error {
		return errors.New("You cannot delete the schema table")
	})
	app.OnRecordUpdateRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return errors.New("You cannot update the schema table")
	})
	app.OnRecordDeleteRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return errors.New("You cannot delete the schema table")
	})
	app.OnRecordCreateRequest(schemaTable).BindFunc(func(e *core.RecordRequestEvent) error {
		return errors.New("You cannot create a record in the schema table")
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

func validateRecordData(app *pocketbase.PocketBase, record *core.Record, schemaTable string) error {

	var collection *core.Collection
	collection, err := app.FindCollectionByNameOrId(schemaTable)
	if err != nil {
		return err
	}

	table := record.Collection().Name

	filter := dbx.Params{
		"table": table,
	}

	tableSchemas, err := app.FindRecordsByFilter(collection, "table = {:table}", "", 0, 0, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No schema to validate against
		}
		return err
	}

	log.Println("Number of schemas found: ", len(tableSchemas))

	for _, schemaRecord := range tableSchemas {

		currentColumn := schemaRecord.GetString("column")
		columnData := record.GetString(currentColumn)

		log.Println("Validating column: ", currentColumn)
		log.Println("Data: ", columnData)

		// Skip validation if the column is empty
		if columnData == "" {
			continue
		}

		schemaContent := schemaRecord.GetString("schema")
		if err := validateJSONSchema(schemaContent); err != nil {
			return err
		}

		schemaLoader := gojsonschema.NewStringLoader(schemaContent)
		dataLoader := gojsonschema.NewStringLoader(columnData)

		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return err
		}

		if !result.Valid() {
			var errMsg string
			for _, desc := range result.Errors() {
				errMsg += desc.String() + "; "
			}
			return apis.NewBadRequestError(fmt.Sprintf("%v validation failed: %s", currentColumn, errMsg), errMsg)
		}
	}

	return nil
}

func validateSchemaTableColumns(app *pocketbase.PocketBase, collection *core.Collection, viewRule *string) (*core.Collection, bool) {

	changed := false

	createOrUpdateCollectionRules(collection, RulesConfig{
		ListRule:   viewRule,
		ViewRule:   viewRule,
		CreateRule: nil,
		DeleteRule: nil,
		UpdateRule: nil,
	}, &changed)

	createOrUpdateTextField(collection, "table", &core.TextField{
		Name:        "table",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: true,
	}, &changed)

	createOrUpdateTextField(collection, "column", &core.TextField{
		Name:        "column",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: true,
	}, &changed)

	createOrUpdateTextField(collection, "hash", &core.TextField{
		Name:        "hash",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: false,
	}, &changed)

	createOrUpdateJSONField(collection, "schema", &core.JSONField{
		Name:        "schema",
		Required:    true,
		Hidden:      false,
		MaxSize:     0,
		Presentable: false,
	}, &changed)

	createOrUpdateAutodateField(collection, "updated", &core.AutodateField{
		Name:        "updated",
		OnCreate:    true,
		OnUpdate:    true,
		Hidden:      false,
		Presentable: false,
	}, &changed)

	createOrUpdateAutodateField(collection, "created", &core.AutodateField{
		Name:        "created",
		OnCreate:    true,
		OnUpdate:    false,
		Hidden:      false,
		Presentable: false},
		&changed)

	return collection, changed

}

func getOrCreateSchemaCollection(app *pocketbase.PocketBase, schemaTable string, viewRule *string) *core.Collection {
	collection, err := app.FindCollectionByNameOrId(schemaTable)
	if err != nil {
		collection = core.NewBaseCollection(schemaTable)
		app.Save(collection)
	}

	collection, changed := validateSchemaTableColumns(app, collection, viewRule)
	if changed {
		app.Save(collection)
	}
	return collection
}
