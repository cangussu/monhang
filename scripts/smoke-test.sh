#!/usr/bin/env bash
# Smoke test for monhang
# This script runs a comprehensive smoke test by:
# 1. Setting up test git repositories
# 2. Building monhang
# 3. Testing workspace sync command
# 4. Testing exec command
# 5. Cleaning up

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEST_DIR="${PROJECT_ROOT}/test-workspace"
MONHANG_BIN="${PROJECT_ROOT}/dist/monhang"

# Error handler
error_exit() {
    echo -e "${RED}✗ Error: $1${NC}" >&2
    exit 1
}

# Success message
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Info message
info() {
    echo -e "${BLUE}→ $1${NC}"
}

# Warning message
warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Cleanup function
cleanup() {
    if [ -d "${TEST_DIR}" ]; then
        info "Cleaning up test directory..."
        rm -rf "${TEST_DIR}"
        success "Cleanup complete"
    fi
}

# Trap cleanup on exit
trap cleanup EXIT

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Monhang Smoke Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Build monhang
info "Step 1: Building monhang..."
cd "${PROJECT_ROOT}"
if ! make build > /dev/null 2>&1; then
    error_exit "Failed to build monhang"
fi

if [ ! -f "${MONHANG_BIN}" ]; then
    error_exit "Binary not found at ${MONHANG_BIN}"
fi
success "Built monhang successfully"
echo ""

# Step 2: Setup test repositories
info "Step 2: Setting up test repositories..."
if ! bash "${SCRIPT_DIR}/setup-test-repos.sh" "${TEST_DIR}" > /dev/null 2>&1; then
    error_exit "Failed to setup test repositories"
fi
success "Test repositories created"
echo ""

# Step 3: Test workspace sync command with JSON
info "Step 3: Testing workspace sync command (JSON config)..."
cd "${TEST_DIR}"

# Run workspace sync command
if ! "${MONHANG_BIN}" workspace sync -f monhang.json > /dev/null 2>&1; then
    error_exit "Workspace sync command failed with JSON config"
fi

# Verify repos were cloned
for repo in lib-utils lib-core lib-network; do
    if [ ! -d "${repo}" ]; then
        error_exit "Repository ${repo} was not cloned"
    fi
    info "  Verified: ${repo}"
done
success "Workspace sync command (JSON) completed successfully"
echo ""

# Step 4: Test workspace sync command with TOML (update existing repos)
info "Step 4: Testing workspace sync command (TOML config - update)..."
cd "${TEST_DIR}"

# Run workspace sync command with TOML (should update existing repos)
if ! "${MONHANG_BIN}" ws sync -f monhang.toml > /dev/null 2>&1; then
    error_exit "Workspace sync command failed with TOML config"
fi

# Verify repos still exist (should be updated, not re-cloned)
for repo in lib-utils lib-core lib-network; do
    if [ ! -d "${repo}" ]; then
        error_exit "Repository ${repo} was not updated (TOML test)"
    fi
    info "  Verified: ${repo}"
done
success "Workspace sync command (TOML - update) completed successfully"
echo ""

# Step 5: Test exec command
info "Step 5: Testing exec command..."
cd "${TEST_DIR}"

# Test basic exec command
if ! "${MONHANG_BIN}" exec -f monhang.json -- pwd > /tmp/exec-test.log 2>&1; then
    error_exit "Exec command failed"
fi

# Verify output contains expected repos
for repo in test-workspace lib-utils lib-core lib-network; do
    if ! grep -q "${repo}" /tmp/exec-test.log; then
        error_exit "Exec output doesn't contain ${repo}"
    fi
    info "  Verified: ${repo} in exec output"
done
success "Exec command completed successfully"
echo ""

# Step 6: Test exec command with parallel
info "Step 6: Testing exec command (parallel)..."
if ! "${MONHANG_BIN}" exec -f monhang.json -p -- echo "test" > /tmp/exec-parallel.log 2>&1; then
    error_exit "Exec command (parallel) failed"
fi

# Check for success message
if ! grep -q "Summary:" /tmp/exec-parallel.log; then
    error_exit "Exec parallel output doesn't contain summary"
fi
success "Exec command (parallel) completed successfully"
echo ""

# Step 7: Test exec with custom script
info "Step 7: Testing exec with custom script..."
if ! "${MONHANG_BIN}" exec -f monhang.json -- ./test-exec.sh > /tmp/exec-script.log 2>&1; then
    error_exit "Exec command with script failed"
fi

# Verify script ran in each repo
for repo in test-workspace lib-utils lib-core lib-network; do
    if ! grep -q "Running test in: ${repo}" /tmp/exec-script.log; then
        error_exit "Script didn't run in ${repo}"
    fi
    info "  Verified: script ran in ${repo}"
done
success "Exec with custom script completed successfully"
echo ""

# Step 8: Test help command
info "Step 8: Testing help command..."
if ! "${MONHANG_BIN}" help > /tmp/help.log 2>&1; then
    error_exit "Help command failed"
fi

# Verify help contains expected commands
for cmd in workspace exec version; do
    if ! grep -q "${cmd}" /tmp/help.log; then
        error_exit "Help doesn't mention ${cmd} command"
    fi
done
success "Help command completed successfully"
echo ""

# Print final summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All smoke tests passed! ✓${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Tests completed:"
echo "  ✓ Build monhang"
echo "  ✓ Setup test repositories"
echo "  ✓ Workspace sync command (JSON config)"
echo "  ✓ Workspace sync command (TOML config - update)"
echo "  ✓ Exec command (sequential)"
echo "  ✓ Exec command (parallel)"
echo "  ✓ Exec command (custom script)"
echo "  ✓ Help command"
echo ""
echo "Monhang is working correctly!"
echo ""
