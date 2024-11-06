package collections

import (
	"embed"
	"encoding/json"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema/*
var content embed.FS

type CollectionPluginConfig struct {
	Enabled                       bool               `mapstructure:"enabled" json:"enabled"`
	RetainUnconfiguredCollections bool               `mapstructure:"retain_unconfigured_collections" json:"retain_unconfigured_collections"`
	FilterPrefix                  string             `mapstructure:"filter_prefix" json:"filter_prefix"`
	Collections                   []CollectionConfig `mapstructure:"collections" json:"collections"`
}

func SetupCollections(app *pocketbase.PocketBase, v *viper.Viper) {

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)
	v.SetDefault("retain_unconfigured_collections", false)
	v.SetDefault("filter_prefix", "_")

	if !v.GetBool("enabled") {
		return
	}

	var genericConfig map[string]interface{}
	err := v.UnmarshalExact(&genericConfig)
	if err != nil {
		panic(err)
	}

	validateSchema(genericConfig)

	pluginConfig := CollectionPluginConfig{}
	err = v.Unmarshal(&pluginConfig)
	if err != nil {
		panic(err)
	}

	// First we need to remove all teh indexes to allow removal of indexed fields.
	for _, collectionConfig := range pluginConfig.Collections {
		collectionConfig.RemoveIndexes(app)
		if collectionConfig.Type == "view" {
			collectionConfig.RemoveCollection(app)
		}
	}

	// First create collections and then create fields to ensure the collections exist prior to creating any reference fields.
	for _, collectionConfig := range pluginConfig.Collections {
		if collectionConfig.Type == "view" {
			continue
		}
		collectionConfig.CreateOrUpdateCollection(app)
	}

	// Create View Collections
	for _, collectionConfig := range pluginConfig.Collections {
		if collectionConfig.Type == "view" {
			collectionConfig.CreateOrUpdateCollection(app)
		}
	}

	for _, collectionConfig := range pluginConfig.Collections {
		collectionConfig.UpdateFields(app)
	}

	pluginConfig.removeUnusedCollections(app)

}

func loadSchemaToJSON(schemaPath string) map[string]interface{} {
	schemaBytes, err := content.ReadFile(schemaPath)
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

func schemaToString(schema map[string]interface{}) string {
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		log.Panicf("Failed to marshal schema: %v", err)
	}
	return string(schemaBytes)
}

func replaceRefWithIdInSchema(schema map[string]interface{}, ref string, id string) {
	for key, value := range schema {
		if key == "$ref" {
			if value == ref {
				schema["$id"] = id
				delete(schema, "$ref")
			}
		}
		// If the value is a map, we need to recurse
		if valueMap, ok := value.(map[string]interface{}); ok {
			replaceRefWithIdInSchema(valueMap, ref, id)
		}
		// If the value is a slice, we need to iterate and recurse
		if valueSlice, ok := value.([]interface{}); ok {
			for _, item := range valueSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					replaceRefWithIdInSchema(itemMap, ref, id)
				}
			}
		}
	}
}

func validateSchema(rawSchema map[string]interface{}) {

	// Load the schema files
	schema_map := loadSchemaToJSON("schema/collections_schema.json")
	schema_fields_map := loadSchemaToJSON("schema/collections_schema_fields.json")

	// Replace the $ref with $id in the schema files
	replaceRefWithIdInSchema(schema_map, "collections_schema_fields.json", "https://raw.githubusercontent.com/qwacko/pocketforge/refs/heads/main/collections/schema/collections_schema_fields.json")

	// Validate the schema
	schemaLoader := gojsonschema.NewStringLoader(schemaToString(schema_map))
	schemaFieldsLoader := gojsonschema.NewStringLoader(schemaToString(schema_fields_map))
	data := gojsonschema.NewStringLoader(schemaToString(rawSchema))

	schema := gojsonschema.NewSchemaLoader()
	schema.AddSchemas(schemaFieldsLoader)
	validated_schema, err := schema.Compile(schemaLoader)

	if err != nil {
		log.Panicf("Failed to compile schema: %v", err)
	}

	result, err := validated_schema.Validate(data)
	if err != nil {
		log.Panicf("Failed to validate collection schema: %v", err)
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		log.Panic("The configuration schema is not valid")
	}

	log.Println("The collection configuration schema is valid")

}

func (config *CollectionPluginConfig) removeUnusedCollections(app *pocketbase.PocketBase) {

	if config.RetainUnconfiguredCollections {
		return
	}

	collections_to_retain := []string{"users"}

	var collections_in_config []string
	var collectionIds_in_config []string
	for _, collectionConfig := range config.Collections {
		collections_in_config = append(collections_in_config, collectionConfig.Name)
		collectionIds_in_config = append(collectionIds_in_config, collectionConfig.ID)
	}

	collections, err := app.FindAllCollections()
	if err != nil {
		log.Panicf("Failed to find collections: %v", err)
	}

	for _, collection := range collections {
		found := false

		// Do not remove collections in the list of collections in the config.
		for _, collectionName := range collections_in_config {
			if collectionName == collection.Name {
				found = true
				break
			}
		}

		// Do not remove collections in the list of items to retain.
		for _, collectionName := range collections_to_retain {
			if collectionName == collection.Name {
				found = true
				break
			}
		}

		// Do not remove collections in the list of items to retain.
		for _, collectionId := range collectionIds_in_config {
			if collectionId == collection.Id {
				found = true
				break
			}
		}

		// Do not remove collections with the filter prefix
		if strings.HasPrefix(collection.Name, config.FilterPrefix) {
			found = true
		}

		// Do not remove system collections
		if collection.System {
			found = true
		}

		if !found {
			collection, err := app.FindCollectionByNameOrId(collection.Name)
			if err != nil {
				log.Panicf("Failed to find collection %s: %v", collection.Name, err)
			}
			app.Delete(collection)
		}
	}
}
