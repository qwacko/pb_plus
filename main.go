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

	// Validate configuration
	jsonschema.BuildSchemaAndValidate(v)

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------
	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{Owner: "qwacko", Repo: "pocketforge"})

	// load jsvm (pb_hooks and pb_migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: v.GetString("settings.migrations_dir"),
		HooksDir:      v.GetString("settings.hooks_dir"),
		HooksWatch:    v.GetBool("settings.hooks_watch"),
		HooksPoolSize: v.GetInt("settings.hooks_pool"),
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  v.GetBool("settings.automigrate"),
		Dir:          v.GetString("settings.migrations_dir"),
	})

	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			publicDir := v.GetString("settings.public_dir")
			indexFallback := v.GetBool("settings.index_fallback")
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir), indexFallback))
			}

			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})

	// Configure schema validation
	validation.ConfigureSchemaValidation(app, v)
	superuser.ConfigureSuperuserOverrides(app, v)
	collections.SetupConfiguredCollections(app, v)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
