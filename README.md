# PB Plus

PB Plus is an extension of the base PocketBase application with additional features that can be turned on and off. Currently, it includes the base PocketBase functionality, and more features will be added in the future.

THe application functions exacly like the base PocketBase application, but with additional features as listed below.

## Additional Features

- [Config From More Locations](#configuration-locations)

## Configuration Locations

Configuration can be read from a toml file (./config.toml), yaml file (./config.yaml), JSON file (./config.json), or environment variables. The configuration is with the following precedence (higher overridea lower):

1. Command Line Flag
2. Environment variables
3. One Of:
   - TOML file
   - YAML file
   - JSON file
     _Note that only the first of the configuration files found will be used._

All the configuration that is available in the base PocketBase application is also available in PB Plus.
