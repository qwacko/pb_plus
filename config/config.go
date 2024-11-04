package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func LoadConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config") // name of config file (without extension)
	v.AddConfigPath(".")      // optionally look for config in the working directory
	v.AutomaticEnv()          // read in environment variables that match

	// Check for config file in multiple formats
	configFiles := []string{"config.toml", "config.yaml", "config.json"}
	for _, configFile := range configFiles {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err == nil {
			break
		}
	}

	// Set default values
	v.SetDefault("hooksDir", defaultHooksDir())
	v.SetDefault("hooksWatch", true)
	v.SetDefault("hooksPool", 15)
	v.SetDefault("migrationsDir", defaultMigrationsDir())
	v.SetDefault("automigrate", true)
	v.SetDefault("publicDir", defaultPublicDir())
	v.SetDefault("indexFallback", true)

	// Bind command line flags
	pflag.String("hooksDir", defaultHooksDir(), "the directory with the JS app hooks")
	pflag.Bool("hooksWatch", true, "auto restart the app on pb_hooks file change")
	pflag.Int("hooksPool", 15, "the total prewarm goja.Runtime instances for the JS app hooks execution")
	pflag.String("migrationsDir", defaultMigrationsDir(), "the directory with the user defined migrations")
	pflag.Bool("automigrate", true, "enable/disable auto migrations")
	pflag.String("publicDir", defaultPublicDir(), "the directory to serve static files")
	pflag.Bool("indexFallback", true, "fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	return v, nil
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}

func defaultHooksDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_hooks"
	}

	return filepath.Join(os.Args[0], "../pb_hooks")
}

func defaultMigrationsDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_migrations"
	}

	return filepath.Join(os.Args[0], "../pb_migrations")
}
