#!/usr/bin/env bash
# Setup test git repositories for monhang testing
# This script creates a set of test git repositories that can be used
# for integration testing and smoke tests.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR_NAME="${1:-test-workspace}"
TEST_DIR="$(pwd)/${TEST_DIR_NAME}"
REPOS_DIR="${TEST_DIR}/repos"

echo -e "${GREEN}Setting up test environment in: ${TEST_DIR}${NC}"

# Clean up if exists
if [ -d "${TEST_DIR}" ]; then
    echo -e "${YELLOW}Removing existing test directory...${NC}"
    rm -rf "${TEST_DIR}"
fi

# Create directory structure
mkdir -p "${REPOS_DIR}"

# Function to create a git repo
create_git_repo() {
    local name=$1
    local desc=$2
    local repo_path="${REPOS_DIR}/${name}.git"

    echo -e "${GREEN}Creating repo: ${name}${NC}"

    # Create bare git repo
    git init --bare "${repo_path}" > /dev/null 2>&1

    # Create a temporary working directory
    local tmp_dir=$(mktemp -d)
    cd "${tmp_dir}"

    # Initialize repo
    git init > /dev/null 2>&1
    git config user.name "Test User"
    git config user.email "test@example.com"
    git config commit.gpgsign false

    # Create some content
    cat > README.md <<EOF
# ${name}

${desc}

This is a test repository for monhang integration testing.
EOF

    # Create a simple build script
    cat > build.sh <<'EOF'
#!/usr/bin/env bash
echo "Building ${name}..."
echo "Build complete!"
EOF
    chmod +x build.sh

    # Create a package file
    cat > package.json <<EOF
{
  "name": "${name}",
  "version": "1.0.0",
  "description": "${desc}"
}
EOF

    # Add and commit
    git add .
    git commit -m "Initial commit: ${name}" > /dev/null 2>&1

    # Tag the version
    git tag -a v1.0.0 -m "Version 1.0.0" > /dev/null 2>&1

    # Get current branch name (supports both master and main)
    local branch=$(git branch --show-current)

    # Push to bare repo
    git remote add origin "file://${repo_path}"
    git push -u origin "${branch}" > /dev/null 2>&1
    git push --tags > /dev/null 2>&1

    # Clean up temp directory
    cd - > /dev/null 2>&1
    rm -rf "${tmp_dir}"

    echo -e "  ${GREEN}✓${NC} Created ${name} at ${repo_path}"
}

# Create test repositories
echo ""
echo "Creating test repositories..."
echo ""

create_git_repo "lib-utils" "Utility library with common functions"
create_git_repo "lib-core" "Core library with base functionality"
create_git_repo "lib-network" "Network communication library"
create_git_repo "app-backend" "Backend application"
create_git_repo "app-frontend" "Frontend application"

# Create monhang configuration
echo ""
echo -e "${GREEN}Creating monhang configuration...${NC}"

cat > "${TEST_DIR}/monhang.json" <<EOF
{
  "name": "test-workspace",
  "version": "1.0.0",
  "repo": "file://${REPOS_DIR}/app-backend.git",

  "repoconfig": {
    "type": "git",
    "base": "file://${REPOS_DIR}/"
  },

  "components": [
    {
      "source": "file://${REPOS_DIR}/lib-core.git?version=v1.0.0&type=git",
      "name": "lib-core",
      "description": "Core library with base functionality",
      "children": [
        {
          "source": "file://${REPOS_DIR}/lib-utils.git?version=v1.0.0&type=git",
          "name": "lib-utils",
          "description": "Utility library with common functions"
        }
      ]
    },
    {
      "source": "file://${REPOS_DIR}/lib-network.git?version=v1.0.0&type=git",
      "name": "lib-network",
      "description": "Network communication library"
    },
    {
      "source": "file://${REPOS_DIR}/app-backend.git?version=v1.0.0&type=git",
      "name": "app-backend",
      "description": "Backend application"
    },
    {
      "source": "file://${REPOS_DIR}/app-frontend.git?version=v1.0.0&type=git",
      "name": "app-frontend",
      "description": "Frontend application"
    }
  ],

  "deps": {
    "build": [
      {
        "name": "lib-utils",
        "version": "v1.0.0",
        "repo": "lib-utils.git"
      },
      {
        "name": "lib-core",
        "version": "v1.0.0",
        "repo": "lib-core.git"
      }
    ],
    "runtime": [
      {
        "name": "lib-network",
        "version": "v1.0.0",
        "repo": "lib-network.git"
      }
    ]
  }
}
EOF

echo -e "  ${GREEN}✓${NC} Created monhang.json"

# Create a TOML version too
cat > "${TEST_DIR}/monhang.toml" <<EOF
name = "test-workspace"
version = "1.0.0"
repo = "file://${REPOS_DIR}/app-backend.git"

[repoconfig]
type = "git"
base = "file://${REPOS_DIR}/"

[[components]]
source = "file://${REPOS_DIR}/lib-core.git?version=v1.0.0&type=git"
name = "lib-core"
description = "Core library with base functionality"

[[components.children]]
source = "file://${REPOS_DIR}/lib-utils.git?version=v1.0.0&type=git"
name = "lib-utils"
description = "Utility library with common functions"

[[components]]
source = "file://${REPOS_DIR}/lib-network.git?version=v1.0.0&type=git"
name = "lib-network"
description = "Network communication library"

[[components]]
source = "file://${REPOS_DIR}/app-backend.git?version=v1.0.0&type=git"
name = "app-backend"
description = "Backend application"

[[components]]
source = "file://${REPOS_DIR}/app-frontend.git?version=v1.0.0&type=git"
name = "app-frontend"
description = "Frontend application"

[[deps.build]]
name = "lib-utils"
version = "v1.0.0"
repo = "lib-utils.git"

[[deps.build]]
name = "lib-core"
version = "v1.0.0"
repo = "lib-core.git"

[[deps.runtime]]
name = "lib-network"
version = "v1.0.0"
repo = "lib-network.git"
EOF

echo -e "  ${GREEN}✓${NC} Created monhang.toml"

# Create test script for use with exec command
cat > "${TEST_DIR}/test-exec.sh" <<'EOF'
#!/usr/bin/env bash
# This script can be used to test the exec command
echo "Running test in: $(basename $(pwd))"
echo "Git status:"
git status --short || echo "Not a git repo"
echo "Files:"
ls -la
EOF
chmod +x "${TEST_DIR}/test-exec.sh"

echo -e "  ${GREEN}✓${NC} Created test-exec.sh"

# Print summary
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Test environment setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Location: ${TEST_DIR}"
echo ""
echo "Repositories created:"
echo "  - lib-utils (file://${REPOS_DIR}/lib-utils.git)"
echo "  - lib-core (file://${REPOS_DIR}/lib-core.git)"
echo "  - lib-network (file://${REPOS_DIR}/lib-network.git)"
echo "  - app-backend (file://${REPOS_DIR}/app-backend.git)"
echo "  - app-frontend (file://${REPOS_DIR}/app-frontend.git)"
echo ""
echo "Configuration files:"
echo "  - monhang.json (JSON format)"
echo "  - monhang.toml (TOML format)"
echo ""
echo "Next steps:"
echo "  1. cd ${TEST_DIR}"
echo "  2. ../dist/monhang boot -f monhang.json"
echo "  3. ../dist/monhang workspace sync -f monhang.json"
echo "  4. ../dist/monhang exec -f monhang.json -- ./test-exec.sh"
echo ""
