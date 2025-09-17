#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "Testing Godo Release Automation"
echo "==============================="

# Test directory
cd "$(dirname "$0")/.."

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_command() {
    local test_name="$1"
    local command="$2"
    local expected_exit_code="${3:-0}"

    echo "Testing: $test_name"

    if eval "$command" > /dev/null 2>&1; then
        actual_exit_code=0
    else
        actual_exit_code=$?
    fi

    if [ "$actual_exit_code" -eq "$expected_exit_code" ]; then
        echo -e "${GREEN}PASS: $test_name${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: $test_name (exit code: $actual_exit_code, expected: $expected_exit_code)${NC}"
        ((TESTS_FAILED++))
    fi
    echo ""
}

# Test make targets exist
echo "Testing Make Targets"
echo "-------------------"

test_command "Make help target" "make help"
test_command "Make changes target" "make changes"
test_command "Make dev-dependencies target" "make dev-dependencies"
test_command "Make test target" "make test"
test_command "Make lint target" "make lint"

# Test scripts exist and are executable
echo "Testing Scripts"
echo "--------------"

test_command "bumpversion.sh exists and executable" "[ -x scripts/bumpversion.sh ]"
test_command "tag.sh exists and executable" "[ -x scripts/tag.sh ]"

# Test script syntax
echo "Testing Script Syntax"
echo "--------------------"

test_command "bumpversion.sh syntax" "bash -n scripts/bumpversion.sh"
test_command "tag.sh syntax" "bash -n scripts/tag.sh"

# Test workflow files
echo "Testing Workflow Files"
echo "---------------------"

test_command "Release workflow exists" "[ -f .github/workflows/release.yml ]"

# Test current version can be extracted
echo "Testing Version Extraction"
echo "-------------------------"

test_command "Can extract version from godo.go" "grep -E '^\s*libraryVersion\s*=' godo.go | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+'"

# Test documentation exists
echo "Testing Documentation"
echo "--------------------"

test_command "Main CONTRIBUTING.md updated" "grep -q 'Releasing' CONTRIBUTING.md"

# Summary
echo "Test Results"
echo "==========="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All tests passed! Release automation is ready to use.${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Install tools: ${BLUE}make dev-dependencies${NC}"
    echo -e "  2. Try it out: ${BLUE}make changes${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Some tests failed. Please fix the issues before using the automation.${NC}"
    exit 1
fi