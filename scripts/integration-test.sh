#!/usr/bin/env bash
# Integration test runner for monhang
# This script can be called from Go tests or run standalone
# It provides a reusable test environment

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEST_DIR="${TEST_DIR:-${PROJECT_ROOT}/test-workspace-integration}"
MONHANG_BIN="${MONHANG_BIN:-${PROJECT_ROOT}/dist/monhang}"

# Parse command line arguments
COMMAND="${1:-help}"
CONFIG_FORMAT="${2:-json}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

usage() {
    cat <<EOF
Usage: $0 <command> [config-format]

Commands:
    setup       - Setup test repositories and configuration
    cleanup     - Remove test environment
    boot        - Run boot command
    exec        - Run exec command with test script
    verify      - Verify repositories were cloned
    all         - Run all tests (setup, boot, verify, exec, cleanup)
    help        - Show this help

Config Format:
    json        - Use JSON configuration (default)
    toml        - Use TOML configuration

Environment Variables:
    TEST_DIR     - Test directory location (default: test-workspace-integration)
    MONHANG_BIN  - Path to monhang binary (default: dist/monhang)

Examples:
    $0 setup                # Setup test environment
    $0 boot json            # Run boot with JSON config
    $0 exec toml            # Run exec with TOML config
    $0 all                  # Run full integration test
EOF
}

cmd_setup() {
    echo -e "${BLUE}Setting up test environment...${NC}"
    bash "${SCRIPT_DIR}/setup-test-repos.sh" "${TEST_DIR}"
}

cmd_cleanup() {
    echo -e "${BLUE}Cleaning up test environment...${NC}"
    if [ -d "${TEST_DIR}" ]; then
        rm -rf "${TEST_DIR}"
        echo -e "${GREEN}✓ Cleanup complete${NC}"
    else
        echo -e "${YELLOW}No test directory to clean${NC}"
    fi
}

cmd_boot() {
    local config_file="monhang.${CONFIG_FORMAT}"
    echo -e "${BLUE}Running boot command (${CONFIG_FORMAT})...${NC}"

    if [ ! -f "${TEST_DIR}/${config_file}" ]; then
        echo -e "${RED}✗ Config file not found: ${config_file}${NC}"
        exit 1
    fi

    cd "${TEST_DIR}"
    "${MONHANG_BIN}" boot -f "${config_file}"
    echo -e "${GREEN}✓ Boot completed${NC}"
}

cmd_verify() {
    echo -e "${BLUE}Verifying repositories...${NC}"
    local all_ok=true

    for repo in lib-utils lib-core lib-network; do
        if [ -d "${TEST_DIR}/${repo}" ]; then
            echo -e "${GREEN}✓ ${repo} exists${NC}"
        else
            echo -e "${RED}✗ ${repo} not found${NC}"
            all_ok=false
        fi
    done

    if [ "$all_ok" = true ]; then
        echo -e "${GREEN}✓ All repositories verified${NC}"
        return 0
    else
        echo -e "${RED}✗ Some repositories missing${NC}"
        return 1
    fi
}

cmd_exec() {
    local config_file="monhang.${CONFIG_FORMAT}"
    echo -e "${BLUE}Running exec command (${CONFIG_FORMAT})...${NC}"

    if [ ! -f "${TEST_DIR}/${config_file}" ]; then
        echo -e "${RED}✗ Config file not found: ${config_file}${NC}"
        exit 1
    fi

    cd "${TEST_DIR}"
    "${MONHANG_BIN}" exec -f "${config_file}" -- pwd
    echo -e "${GREEN}✓ Exec completed${NC}"
}

cmd_all() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Running full integration test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    cmd_setup
    echo ""

    cmd_boot
    echo ""

    cmd_verify
    echo ""

    cmd_exec
    echo ""

    cmd_cleanup
    echo ""

    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}All integration tests passed!${NC}"
    echo -e "${GREEN}========================================${NC}"
}

# Main command dispatcher
case "$COMMAND" in
    setup)
        cmd_setup
        ;;
    cleanup)
        cmd_cleanup
        ;;
    boot)
        cmd_boot
        ;;
    verify)
        cmd_verify
        ;;
    exec)
        cmd_exec
        ;;
    all)
        cmd_all
        ;;
    help)
        usage
        ;;
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        echo ""
        usage
        exit 1
        ;;
esac
