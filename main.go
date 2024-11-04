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

	config "pb_plus/config"
	"pb_plus/validation" // Import the new validation package
)

func main() {
	app := pocketbase.New()

	// Load configuration
	v, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{Owner: "qwacko", Repo: "pb_plus"})

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

	// Configure schema validation
	validation.ConfigureSchemaValidation(app, v)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
