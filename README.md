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

## Git Operations Across Repositories

Monhang provides a `git` command to perform git operations across all repositories defined in your configuration file. This is particularly useful for managing multiple related components.

### Check Status

Show the git status for all repositories:

```sh
monhang git status
monhang git -i status     # interactive mode with live UI
```

Example output:
```
Repository                     Branch               Commit     Status
---------------------------------------------------------------------------------
repo1                          feat/git-commands    2795659    2 changes
repo2                          feat/git-commands    57fa960    clean
```

### Pull Updates

Pull the latest changes for all repositories:

```sh
monhang git pull
monhang git -p pull       # run in parallel
```

### Fetch Updates

Fetch updates from remotes without merging:

```sh
monhang git fetch
monhang git -p fetch      # run in parallel
```

### Checkout Branch

Checkout a specific branch across all repositories:

```sh
monhang git checkout main
monhang git -i checkout develop    # interactive mode
```

### Create and Checkout Branch

Create a new branch and check it out across all repositories:

```sh
monhang git branch feat/new-feature
monhang git -p branch feat/new-feature    # run in parallel
```

### Options

All git subcommands support the following options:

- `-f <file>`: Specify configuration file (default: `./monhang.json`)
- `-i`: Interactive mode with bubbletea UI for live progress
- `-p`: Run operations in parallel for faster execution

**Note:** Git operations only affect repositories that have already been cloned locally (via `monhang boot`).

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
