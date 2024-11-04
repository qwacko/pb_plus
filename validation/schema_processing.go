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
	"github.com/xeipuuv/gojsonschema"
)

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

	for _, schemaRecord := range tableSchemas {

		currentColumn := schemaRecord.GetString("column")
		columnData := record.GetString(currentColumn)

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
