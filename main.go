package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/xeipuuv/gojsonschema"

	"pocketforge/collections"
	"pocketforge/config" //Import the new config package
	"pocketforge/jsonschema"
	"pocketforge/superuser"
	"pocketforge/validation" // Import the new validation package
)

func main() {

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDev: false,
	})

	// Load configuration
	v, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{Owner: "qwacko", Repo: "pocketforge"})

	// load jsvm (pb_hooks and pb_migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: v.GetString("migrationsDir"),
		HooksDir:      v.GetString("hooksDir"),
		HooksWatch:    v.GetBool("hooksWatch"),
		HooksPoolSize: v.GetInt("hooksPool"),
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  v.GetBool("automigrate"),
		Dir:          v.GetString("migrationsDir"),
	})

	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			publicDir := v.GetString("publicDir")
			indexFallback := v.GetBool("indexFallback")
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir), indexFallback))
			}

			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})

	schema, err := jsonschema.BuildSchema()
	if err != nil {
		log.Fatalf("Failed to build schema: %v", err)
	}

	var genericConfig map[string]interface{}
	err = v.UnmarshalExact(&genericConfig)
	if err != nil {
		panic(err)
	}
	data := gojsonschema.NewStringLoader(jsonschema.SchemaToString(genericConfig))

	result, err := schema.Validate(data)
	if err != nil {
		log.Panicf("Failed to validate collection schema: %v", err)
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		log.Panic("The configuration schema is not valid")
	}

	// Configure schema validation
	validation.ConfigureSchemaValidation(app, v)
	superuser.ConfigureSuperuserOverrides(app, v)
	collections.SetupConfiguredCollections(app, v)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
