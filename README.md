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

## Logging and Debug Mode

Monhang uses [zerolog](https://github.com/rs/zerolog) for structured logging. By default, the log level is set to `INFO`, which provides essential information about operations. You can enable debug mode to see detailed logging information, which is helpful for troubleshooting and understanding what monhang is doing.

### Enabling Debug Mode

There are two ways to enable debug mode:

#### 1. Environment Variable

Set the `MONHANG_DEBUG` environment variable to `true`:

```sh
export MONHANG_DEBUG=true
monhang boot -f monhang.json
```

Or run it inline:

```sh
MONHANG_DEBUG=true monhang boot -f monhang.json
```

#### 2. Command-Line Flag

Use the `--debug` flag:

```sh
monhang --debug boot -f monhang.json
monhang --debug git status
monhang --debug exec -- make build
```

**Note:** The `--debug` flag must come **before** the subcommand (e.g., `boot`, `git`, `exec`).

### What Debug Mode Shows

When debug mode is enabled, you'll see detailed information about:

- Configuration file parsing
- Dependency graph processing and sorting
- Git command execution (commands, arguments, output)
- Repository operations and timing
- Command execution details (start, completion, duration)
- Error context and stack traces

### Example Output

Normal mode (INFO level):
```
2:04:05PM INF Starting bootstrap process config=./monhang.json component=component
2:04:05PM INF Project loaded project=top-app version=1.0.3 component=component
2:04:05PM INF Fetching dependencies... component=component
```

Debug mode:
```
2:04:05PM DBG Debug mode enabled
2:04:05PM INF Starting bootstrap process config=./monhang.json component=component
2:04:05PM DBG Parsing project file extension=.json filename=./monhang.json component=component
2:04:05PM INF Project loaded project=top-app version=1.0.3 component=component
2:04:05PM DBG Processing project dependencies project=top-app component=component
2:04:05PM DBG Processing dependency dependency=lib1 type=build component=component
2:04:05PM DBG Adding toplevel repoconfig to dependency base=git@github.com:monhang/ dependency=lib1 component=component
2:04:05PM DBG Finished processing dependencies project=top-app total_deps=2 component=component
2:04:05PM DBG Sorting project dependencies project=top-app component=component
2:04:05PM DBG Dependencies sorted project=top-app sorted_count=3 component=component
2:04:05PM INF Fetching dependencies... component=component
2:04:05PM INF Fetching component component=lib1 version=v1.0.0 component=component
2:04:05PM INF Executing git command args=["clone" "git@github.com:monhang/lib1.git" "lib1"] component=component
```

### Log Format

Logs are displayed in a human-readable console format with:
- **Timestamp**: Time of the log entry (HH:MM:SS format)
- **Level**: Log level (DBG, INF, WRN, ERR, FAT)
- **Message**: Descriptive message
- **Fields**: Structured contextual data (component, repo, duration, etc.)
- **Color coding**: Different colors for different log levels (in supported terminals)

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
