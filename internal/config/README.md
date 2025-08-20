# Internal Config Package

This package contains all configuration implementation details that are not part of the public API.

## Files

- `loader.go` - TOML parsing, file reading, defaults setting
- `validator.go` - Comprehensive validation logic, error reporting  
- `resolver.go` - Environment variable resolution, config merging
- `reloader.go` - File watching, hot reloading, change detection

## Purpose

These implementation details were moved from the core package to keep the public API clean and minimal. The core package now only exposes essential configuration types and the LoadConfig function.