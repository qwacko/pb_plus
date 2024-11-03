# PB Plus

PB Plus is an extension of the base PocketBase application with additional features that can be turned on and off. Currently, it includes the base PocketBase functionality, and more features will be added in the future.

## Features

- Base PocketBase functionality
- JavaScript VM for app hooks and migrations
- Migration command with JavaScript templates
- Static file serving from a specified directory

## Optional Plugin Flags

The following flags can be used to configure the application:

- `--hooksDir`: The directory with the JS app hooks.
- `--hooksWatch`: Auto restart the app on pb_hooks file change (default: true).
- `--hooksPool`: The total prewarm goja.Runtime instances for the JS app hooks execution (default: 15).
- `--migrationsDir`: The directory with the user-defined migrations.
- `--automigrate`: Enable/disable auto migrations (default: true).
- `--publicDir`: The directory to serve static files (default: "./pb_public").
- `--indexFallback`: Fallback the request to index.html on missing static path (default: true).

## Usage

To start the application, run:
