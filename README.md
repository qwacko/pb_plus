# Pocketforge

Pocketforge is an extension of the base PocketBase application with additional features that can be turned on and off. Currently, it includes the base PocketBase functionality, and more features will be added in the future.

The application functions exactly like the base PocketBase application (other than using the `pocketforge` command rather than `pocketbase`), but with additional features as listed below.

# Additional Features

- [Config From More Locations](#configuration-locations)
- [Automatic Updates](#automatic-updates)
- [JSON Schema Validation](#json-schema-validation) - Allows validation of json columns against schemas.
- [Superuser Management](#superuser-management)
- Collection Configuration From File - **Future**
- Chat GPT / Open Router Integration - **Future**
- Fetch Requests - **Future**
- Settings Automatic Loading - **Future**
- Schema Summary Endpoint / Diagram / Raw SQL (For AI to help writing view queries) - **Future**

# Configuration Locations

Configuration can be read from a TOML file, YAML file, JSON file, or environment variables. The configuration is with the following precedence (first in the list overrides later):

1. Environment variables
2. One Of the following (only the first one found will be used):

   - TOML file (./config.toml)
   - YAML file (./config.yaml)
   - JSON file (./config.json)

Configuration parameters from pocketbase are supported in pocketforge,however they are in a `settings` tree (i.e. `settings.automigrate` instead of `automigrate`).

## Example Configuration Files

Example configuration files are included in the GitHub repository to help you get started. You can find them in the `examples` directory:

- [config.toml](examples/example_config.toml)
- [config.yaml](examples/example_config.yaml)
- [config.json](examples/example_config.json)
- [.env](examples/example_env.env)

Feel free to copy and modify these files to suit your needs.

# JSON Schema Validation

This feature allows you to store JSON schema files in a directory and validate JSON columns in your database against these schemas. This feature is disabled by default and can be enabled by setting the `enabled` configuration parameter to `true`.

The schema information is loaded into a pocketbase collection that is automatically created when the application starts. The collection is named `_schema` by default, but you can change the name by setting the `table` configuration parameter. This table is prevented from anyone (including superusers) from editing the table structure or the data in the table.

The `viewRule` configuration parameter can be used to adjust who can see the schema information. This defaults to superusers only. Making the schema information available to users of the API may be useful as it allows the user to ensure they are correctly providing the data that is expected.

Note that the data is only validated on record creation or update, so incorrectly stored data will be served up.

> **Warning:** JSON schema validation is done on every update. If the schema changes and the stored data is invalid against the new schema, it is not possible to update other fields in the record without also updating the JSON field to be valid against the new schema.

## Configuration Parameters

_Note that all configuration parameters are in the `validation` section of the configuration file. If the `validation` section is not present, then the `_schema` table is not configured._

- `enabled` (bool): Enable or disable JSON schema validation. Default is `true`.
- `schemaDir` (string): Directory where JSON schema files are stored. Default is `./pb_schema`.
- `table` (string): Database table used for storing schema information. Default is `_schema`.
- `viewRule` (bool): Restrict schema viewing to authorized users only. Default is `true`.
- `schema` (array): An array of schema objects. Each object has the following parameters:
  - `filename` (string): File name of the schema file.
  - `table` (string): Table name to validate against.
  - `column` (string): Column name to validate against.

## Example Configuration

```yaml
validation:
  enabled: true
  schema_dir: "./pb_schema"
  table: "_schema"
  viewRule: "@request.auth.id != ''"
  schema:
    - table: "testtable"
      column: "testcolumn"
      filename: "schema.json"
```

# Superuser Management

Superusers can be automatically added to the system, and certain actions from superusers prevented using the configuration file. You can define superuser accounts and specify permissions to restrict actions such as creating, editing, or deleting collections and records.

## Adding Superusers

Superusers can be automaticall added (and passwords automatically updated) by specifying their email and password in the configuration file:

```yaml
superusers:
  - email: "superuser1@example.com"
    password: "password1"
  - email: "superuser2@example.com"
    password: "password2"
```

> **Warning:** The password is stored in plain text in the configuration file. Make sure to secure the configuration file or remove the password after adding the superusers. Superusers that exist in the database but not in the configuration file will **not** be automatically removed (or have their passwords automatically updated).

## Preventing Actions for Superusers

You can prevent superusers from performing certain actions by specifying permissions in the configuration file (`config.yaml`):

```yaml
superuser:
  collections:
    - name: "testtable"
      preventCollectionCreate: true
      preventCollectionUpdate: true
      preventCollectionDelete: true
      preventRecordCreate: false
      preventRecordUpdate: true
      preventRecordDelete: true
```
