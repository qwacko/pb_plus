package jsonschema

import (
	"embed"
	"encoding/json"
	"log"
)

func LoadSchemaToJSON(schemaPath string, source *embed.FS) map[string]interface{} {
	schemaBytes, err := source.ReadFile(schemaPath)
	if err != nil {
		log.Panicf("Failed to read schema: %v", err)
	}
	var schema map[string]interface{}
	err = json.Unmarshal(schemaBytes, &schema)
	if err != nil {
		log.Panicf("Failed to unmarshal schema: %v", err)
	}
	return schema
}

func SchemaToString(schema map[string]interface{}) string {
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		log.Panicf("Failed to marshal schema: %v", err)
	}
	return string(schemaBytes)
}

func ReplaceRefWithIdInSchema(schema map[string]interface{}, ref string, id string) {
	for key, value := range schema {
		if key == "$ref" {
			if value == ref {
				schema["$id"] = id
				delete(schema, "$ref")
			}
		}
		// If the value is a map, we need to recurse
		if valueMap, ok := value.(map[string]interface{}); ok {
			ReplaceRefWithIdInSchema(valueMap, ref, id)
		}
		// If the value is a slice, we need to iterate and recurse
		if valueSlice, ok := value.([]interface{}); ok {
			for _, item := range valueSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					ReplaceRefWithIdInSchema(itemMap, ref, id)
				}
			}
		}
	}
}
