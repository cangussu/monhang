# monhang component management tool

Monhang is a tool that takes the pain out of component management. This is a
development version and all commands can change without notice.

For now, only Git is supported as a component repository.


## Bootstraping a workspace

To fetch a component with all its dependencies, just type:

```sh
monhang boot -f monhang.json
```

Or using a TOML configuration file:

```sh
monhang boot -f monhang.toml
```

This will create a workspace in the current directory with all needed components
as described by the configuration file.

You can also bootstrap from a git URL:

```sh
monhang boot -f git@github.com:cangussu/monhang.git
```

This will clone the repository, process it's monhang.json file and bootstrap the
workspace.

## Configuration file

A configuration file describes a component and also its dependencies. Configuration
files can be written in either **JSON** or **TOML** format. The format is automatically
detected based on the file extension (`.json` or `.toml`).

A component has the following basic information:

- **name**: the component identification
- **version**: the version that will be checked out
- **repository**: the git clone argument

**JSON format:**
```json
{
  "name": "top-app",
  "version": "1.0.3",
  "repo": "git@github.com:cangussu/monhang.git"
}
```

**TOML format:**
```toml
name = "top-app"
version = "1.0.3"
repo = "git@github.com:cangussu/monhang.git"
```

### Dependencies

The dependency object defines three types of dependency: *build*, *runtime* and
*install* time. The following example shows two dependencies:

**JSON format:**
```json
{
  "deps" : {
    "build": [
      {
        "name": "lib1",
        "version": "v1.0.0",
        "repo": "git@github.com:monhang/examples.git"
      },
      {
        "name": "lib2",
        "version": "v2.0.2",
        "repo": "git@github.com:monhang/examples.git"
      }
    ]
  }
}
```

**TOML format:**
```toml
[[deps.build]]
name = "lib1"
version = "v1.0.0"
repo = "git@github.com:monhang/examples.git"

[[deps.build]]
name = "lib2"
version = "v2.0.2"
repo = "git@github.com:monhang/examples.git"
```
