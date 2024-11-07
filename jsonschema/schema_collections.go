package jsonschema

import (
	"embed"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema/**
var content embed.FS

func BuildSchema() (*gojsonschema.Schema, error) {

	resultPrefix := "https://raw.githubusercontent.com/qwacko/pocketforge/refs/heads/main/jsonschema/schema/"

	collections_schema_location := "collections/collections_schema.json"
	collections_schema_fields_location := "collections/collections_schema_fields.json"
	collections_schema_fields_filename := "collections_schema_fields.json"
	superusers_schema_location := "superuser/superuser_schema.json"
	validation_schema_location := "validation/validation_schema.json"

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
				},
			},
			{
				Filename:     "schema/" + collections_schema_fields_location,
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
		},
	}

	validator, err := schemaConfig.BuildCombinedSchema(&content)
	if err != nil {
		return nil, err
	}

	return validator, nil

}
