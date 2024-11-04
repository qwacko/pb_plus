package validation

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
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

	if !v.GetBool("enabled") {
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		schemaDir := v.GetString("schemaDir")
		schemaTable := v.GetString("table")

		var configs []SchemaConfig
		if err := v.UnmarshalKey("schema", &configs); err != nil {
			log.Fatalf("Error unmarshalling schema configurations: %v", err)
		}

		log.Printf("Schema validation enabled, using schemaDir: %s, table: %s", schemaDir, schemaTable)

		for _, config := range configs {
			schemaPath := schemaDir + "/" + config.Filename
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				//Throw an error if the schema file does not exist as we don't want
				//to continue without the schema file
				return err
			}

			// Add your schema validation logic here
			// For example, you can load the schema and set up validation hooks
			log.Printf("Configuring schema validation for DB: %s, Column: %s, using schema: %s", config.Table, config.Column, schemaPath)
		}

		return e.Next()
	})

}
