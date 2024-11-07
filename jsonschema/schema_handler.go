package jsonschema

import (
	"embed"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaReplacement struct {
	Ref string
	Id  string
}

type SingleSchema struct {
	Filename     string
	Replacements []SchemaReplacement
}

type SchemaDefinition struct {
	CoreSchema  SingleSchema
	OtherSchema []SingleSchema
}

func (schema SingleSchema) LoadAndReplace(source *embed.FS) string {

	json_data := LoadSchemaToJSON(schema.Filename, source)

	for _, replacement := range schema.Replacements {
		ReplaceRefWithIdInSchema(json_data, replacement.Ref, replacement.Id)
	}

	return SchemaToString(json_data)
}

// BuildCombinedSchema compiles and returns a combined JSON schema from the core schema and other schemas.
//
// Parameters:
// - content: A pointer to an embedded filesystem containing the schema files.
//
// Returns:
//
// - A pointer to a compiled gojsonschema.Schema.
//
// - An error, if any.
//
// The function iterates over the OtherSchema slice in the SchemaDefinition struct,
// loads and replaces each schema, and adds them to the schema loader. It then compiles
// the core schema and returns the validated schema.
func (def *SchemaDefinition) BuildCombinedSchema(content *embed.FS) (*gojsonschema.Schema, error) {

	schema := gojsonschema.NewSchemaLoader()

	for _, singleSchema := range def.OtherSchema {
		schemaString := singleSchema.LoadAndReplace(content)
		schema.AddSchemas(gojsonschema.NewStringLoader(schemaString))
	}

	validated_schema, err := schema.Compile(gojsonschema.NewStringLoader(def.CoreSchema.LoadAndReplace(content)))

	if err != nil {
		return nil, err
	}

	return validated_schema, nil

}
