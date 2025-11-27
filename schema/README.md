# Monhang Configuration Schema

This directory contains the JSON Schema definition for monhang configuration files.

## Files

- `monhang-manifest.schema.json` - JSON Schema (Draft-07) for validating configuration files
- `monhang-manifest.example.toml` - Example TOML configuration file with documentation

## Schema Overview

The schema validates both JSON and TOML configuration files used by monhang. It ensures:

- **Required fields**: Every component must have a `name`
- **Name format**: Names can only contain letters, numbers, underscores, and hyphens
- **Source URLs**: Must use valid schemes (git://, https://, http://, file://, ssh://)
- **Version format**: Must follow semantic versioning (e.g., v1.0.0, 1.0.0, 1.0.0-beta)
- **No typos**: Additional properties not defined in the schema are rejected
- **Recursive validation**: Nested components are validated with the same rules

## Configuration Fields

### Required Fields

- `name` (string): Component name, used as directory name
  - Pattern: `^[a-zA-Z0-9_-]+$`
  - Minimum length: 1

### Optional Fields

- `version` (string): Component version
  - Pattern: `^v?[0-9]+\.[0-9]+\.[0-9]+(-.+)?$`
  - Examples: `v1.0.0`, `1.0.0`, `1.0.0-beta.1`

- `source` (string): Repository URL with optional query parameters
  - Pattern: `^(git|https?|file|ssh)://.+`
  - Query parameters:
    - `version`: Git tag, branch, or commit to checkout
    - `type`: Repository type (default: "git")
  - Examples:
    - `git://github.com/org/repo.git?version=v1.0.0`
    - `https://github.com/org/repo.git?version=main`
    - `file:///path/to/local/repo.git?version=v2.0.0`

- `description` (string): Human-readable component description

- `components` (array): Child components (recursive)
  - Each item follows the same schema structure

## IDE Support

### VS Code

The schema is automatically detected for files matching these patterns:
- `**/monhang.json`
- `**/monhang-manifest.json`
- `**/.monhang.json`

VS Code will provide:
- Autocomplete for field names
- Validation errors inline
- Hover documentation
- JSON formatting

### IntelliJ / JetBrains IDEs

Configure JSON Schema mapping:
1. Go to: Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema Mappings
2. Add new mapping:
   - Schema file: `<project>/schema/monhang-manifest.schema.json`
   - Schema version: JSON Schema version 7
   - File path patterns: `monhang.json`, `monhang-manifest.json`

## Validation

All configuration files are automatically validated when parsed by monhang. If validation fails, you'll see a detailed error message explaining what's wrong and how to fix it.

### Example Validation Errors

**Missing required field:**
```
configuration validation failed:
  - Field 'root': missing properties: 'name'
```

**Invalid source URL:**
```
configuration validation failed:
  - Field '/source': must be a valid URL starting with git://, https://, file://, or ssh://
```

**Invalid name format:**
```
configuration validation failed:
  - Field '/name': must contain only letters, numbers, underscores, and hyphens
```

**Invalid version format:**
```
configuration validation failed:
  - Field '/version': must follow semantic versioning format (e.g., v1.0.0 or 1.0.0)
```

## Testing Your Configuration

You can test your configuration file by running any monhang command that loads it:

```bash
# Test with workspace sync (dry run would be ideal but sync is current best option)
monhang workspace list -f monhang.json

# Test with exec command (won't execute, just validates config)
monhang exec -f monhang.json echo "test"
```

If the configuration is invalid, monhang will report validation errors immediately.

## JSON Schema Resources

- [JSON Schema Official Site](https://json-schema.org/)
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/)
- [JSON Schema Draft-07 Specification](https://json-schema.org/draft-07/json-schema-release-notes.html)

## Contributing

If you find issues with the schema or have suggestions for improvements, please open an issue or submit a pull request.
