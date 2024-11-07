package config

import (
	"os"
	"path/filepath"
	"strings"

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
	v.SetDefault("settings.hooks_dir", defaultHooksDir())
	v.SetDefault("settings.hooks_watch", true)
	v.SetDefault("settings.hooks_pool", 15)
	v.SetDefault("settings.migrations_dir", defaultMigrationsDir())
	v.SetDefault("settings.automigrate", true)
	v.SetDefault("settings.public_dir", defaultPublicDir())
	v.SetDefault("settings.index_fallback", true)

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
