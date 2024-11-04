# PB Plus

PB Plus is an extension of the base PocketBase application with additional features that can be turned on and off. Currently, it includes the base PocketBase functionality, and more features will be added in the future.

The application functions exactly like the base PocketBase application, but with additional features as listed below.

# Additional Features

- [Config From More Locations](#configuration-locations)

# Configuration Locations

Configuration can be read from a TOML file (./config.toml), YAML file (./config.yaml), JSON file (./config.json), or environment variables. The configuration is with the following precedence (higher overrides lower):

1. Command Line Flag
2. Environment variables
3. One Of:

   - TOML file
   - YAML file
   - JSON file

     _Note that only the first of the configuration files found will be used._

All the configuration that is available in the base PocketBase application is also available in PB Plus.

## Example Configuration Files

Example configuration files are included in the GitHub repository to help you get started. You can find them in the `examples` directory:

- [config.toml](examples/example_config.toml)
- [config.yaml](examples/example_config.yaml)
- [config.json](examples/exampleconfig.json)
- [.env](examples/example_env.env)

Feel free to copy and modify these files to suit your needs.
