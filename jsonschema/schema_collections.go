package jsonschema

import (
	"embed"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema/**
var content embed.FS

func BuildSchema() (*gojsonschema.Schema, error) {

	resultPrefix := "https://raw.githubusercontent.com/qwacko/pocketforge/refs/heads/main/jsonschema/"

	collections_schema_location := "schema/collections/collections_schema.json"
	collections_schema_fields_location := "schema/collections/collections_schema_fields.json"
	collections_schema_fields_filename := "collections_schema_fields.json"

	var schemaConfig = SchemaDefinition{
		CoreSchema: SingleSchema{
			Filename: "schema/config_schema.json",
			Replacements: []SchemaReplacement{
				{
					Ref: collections_schema_location,
					Id:  resultPrefix + collections_schema_location,
				},
			},
		},
		OtherSchema: []SingleSchema{
			{
				Filename: collections_schema_location,
				Replacements: []SchemaReplacement{
					{
						Ref: collections_schema_fields_filename,
						Id:  resultPrefix + collections_schema_fields_location,
					},
				},
			},
			{
				Filename:     collections_schema_fields_location,
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
