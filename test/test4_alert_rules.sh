#!/bin/bash

# Test Prometheus alerting rules using promtool
# This validates that alert rules fire correctly given specific metric values.

set -e

TEST_DIR=$(dirname "$0")
REPO_ROOT="${TEST_DIR}/.."

echo "test4_alert_rules: Testing Prometheus alert rules with promtool"

cd "$REPO_ROOT"

if ! command -v promtool &> /dev/null; then
    echo "ERROR: promtool not found. Install with: apt install prometheus"
    exit 1
fi

# Count number of test cases
test_count=$(grep -c "eval_time:" test/alert_rules_test.yaml)
echo "test4_alert_rules: Running $test_count test cases..."

promtool test rules test/alert_rules_test.yaml

echo "test4_alert_rules: All $test_count test cases passed"
