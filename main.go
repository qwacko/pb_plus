package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	app := pocketbase.New()

	// ---------------------------------------------------------------
	// Load configuration using Viper:
	// ---------------------------------------------------------------

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AutomaticEnv()          // read in environment variables that match

	// Check for config file in multiple formats
	configFiles := []string{"config.toml", "config.yaml", "config.json"}
	for _, configFile := range configFiles {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err == nil {
			break
		}
	}

	// Set default values
	viper.SetDefault("hooksDir", "")
	viper.SetDefault("hooksWatch", true)
	viper.SetDefault("hooksPool", 15)
	viper.SetDefault("migrationsDir", "")
	viper.SetDefault("automigrate", true)
	viper.SetDefault("publicDir", defaultPublicDir())
	viper.SetDefault("indexFallback", true)

	// Bind command line flags
	pflag.String("hooksDir", "", "the directory with the JS app hooks")
	pflag.Bool("hooksWatch", true, "auto restart the app on pb_hooks file change")
	pflag.Int("hooksPool", 15, "the total prewarm goja.Runtime instances for the JS app hooks execution")
	pflag.String("migrationsDir", "", "the directory with the user defined migrations")
	pflag.Bool("automigrate", true, "enable/disable auto migrations")
	pflag.String("publicDir", defaultPublicDir(), "the directory to serve static files")
	pflag.Bool("indexFallback", true, "fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// load jsvm (pb_hooks and pb_migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: viper.GetString("migrationsDir"),
		HooksDir:      viper.GetString("hooksDir"),
		HooksWatch:    viper.GetBool("hooksWatch"),
		HooksPoolSize: viper.GetInt("hooksPool"),
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  viper.GetBool("automigrate"),
		Dir:          viper.GetString("migrationsDir"),
	})

	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			publicDir := viper.GetString("publicDir")
			indexFallback := viper.GetBool("indexFallback")
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir), indexFallback))
			}

			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
