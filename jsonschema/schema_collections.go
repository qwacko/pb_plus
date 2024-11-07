package jsonschema

import (
	"embed"
	"log"

	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema/**
var content embed.FS

func BuildSchema() (*gojsonschema.Schema, error) {

	resultPrefix := "https://raw.githubusercontent.com/qwacko/pocketforge/refs/heads/main/jsonschema/schema/"

	collections_schema_location := "collections/collections_schema.json"
	collections_schema_fields_location := "collections/collections_schema_fields.json"
	collections_schema_fields_filename := "collections_schema_fields.json"
	collections_schema_auth_location := "collections/collections_schema_auth.json"
	collections_schema_auth_filename := "collections_schema_auth.json"
	superusers_schema_location := "superuser/superuser_schema.json"
	validation_schema_location := "validation/validation_schema.json"
	settings_schema_location := "settings/settings_schema.json"

	var schemaConfig = SchemaDefinition{
		CoreSchema: SingleSchema{
			Filename: "schema/config_schema.json",
			Replacements: []SchemaReplacement{
				{
					Ref: collections_schema_location,
					Id:  resultPrefix + collections_schema_location,
				},
				{
					Ref: superusers_schema_location,
					Id:  resultPrefix + superusers_schema_location,
				},
				{
					Ref: validation_schema_location,
					Id:  resultPrefix + validation_schema_location,
				},
				{
					Ref: settings_schema_location,
					Id:  resultPrefix + settings_schema_location,
				},
			},
		},
		OtherSchema: []SingleSchema{
			{
				Filename: "schema/" + collections_schema_location,
				Replacements: []SchemaReplacement{
					{
						Ref: collections_schema_fields_filename,
						Id:  resultPrefix + collections_schema_fields_location,
					},
					{
						Ref: collections_schema_auth_filename,
						Id:  resultPrefix + collections_schema_auth_location,
					},
				},
			},
			{
				Filename:     "schema/" + collections_schema_fields_location,
				Replacements: []SchemaReplacement{},
			},
			{
				Filename:     "schema/" + collections_schema_auth_location,
				Replacements: []SchemaReplacement{},
			},
			{
				Filename:     "schema/" + superusers_schema_location,
				Replacements: []SchemaReplacement{},
			},
			{
				Filename:     "schema/" + validation_schema_location,
				Replacements: []SchemaReplacement{},
			},
			{
				Filename:     "schema/" + settings_schema_location,
				Replacements: []SchemaReplacement{},
			},
		},
	}

	validator, err := schemaConfig.BuildCombinedSchema(&content)
	if err != nil {
		return nil, err
	}

	return validator, nil

}

func BuildSchemaAndValidate(v *viper.Viper) {
	schema, err := BuildSchema()
	if err != nil {
		log.Fatalf("Failed to build schema: %v", err)
	}

	var genericConfig map[string]interface{}
	err = v.UnmarshalExact(&genericConfig)
	if err != nil {
		panic(err)
	}
	data := gojsonschema.NewStringLoader(SchemaToString(genericConfig))

	result, err := schema.Validate(data)
	if err != nil {
		log.Panicf("Failed to validate collection schema: %v", err)
	}

	if !result.Valid() {
		log.Println("The configuration schema is not valid")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		log.Panic("The configuration schema is not valid")
	}

	log.Println("The configuration schema is valid")
}
