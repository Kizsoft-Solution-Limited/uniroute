#!/bin/bash
# Script to move test files to new structure
# This script copies files and updates package names

set -e

# Function to move and update a test file
move_test_file() {
    local old_path=$1
    local new_path=$2
    local is_integration=$3
    
    if [ ! -f "$old_path" ]; then
        echo "Warning: $old_path not found, skipping"
        return
    fi
    
    # Create directory if it doesn't exist
    mkdir -p "$(dirname "$new_path")"
    
    # Get the package name from the old file
    old_package=$(grep "^package " "$old_path" | head -1 | sed 's/^package //')
    
    # Determine new package name
    if [[ "$new_path" == *"/unit/"* ]]; then
        # Unit test: package_name_test
        new_package="${old_package}_test"
    elif [[ "$new_path" == *"/integration/"* ]]; then
        # Integration test: just "integration" or keep original
        new_package="integration"
    else
        new_package="${old_package}_test"
    fi
    
    # Copy file
    cp "$old_path" "$new_path"
    
    # Update package name
    sed -i.bak "s/^package $old_package$/package $new_package/" "$new_path"
    rm -f "${new_path}.bak"
    
    # Add testutil import if it's an integration test
    if [ "$is_integration" = "true" ]; then
        if ! grep -q "github.com/Kizsoft-Solution-Limited/uniroute/tests/testutil" "$new_path"; then
            # Add import after existing imports
            sed -i.bak '/^import (/,/^)$/{
                /^)$/i\
	"github.com/Kizsoft-Solution-Limited/uniroute/tests/testutil"
            }' "$new_path"
            rm -f "${new_path}.bak"
        fi
    fi
    
    echo "Moved: $old_path -> $new_path (package: $old_package -> $new_package)"
}

# Move unit tests
move_test_file "internal/security/apikey_test.go" "tests/unit/security/apikey_test.go" false
move_test_file "internal/security/apikey_v2_test.go" "tests/unit/security/apikey_v2_test.go" false
move_test_file "internal/security/ratelimit_test.go" "tests/unit/security/ratelimit_test.go" false
move_test_file "internal/security/provider_key_service_test.go" "tests/unit/security/provider_key_service_test.go" false
move_test_file "internal/gateway/router_test.go" "tests/unit/gateway/router_test.go" false
move_test_file "internal/gateway/cost_calculator_test.go" "tests/unit/gateway/cost_calculator_test.go" false
move_test_file "internal/gateway/latency_tracker_test.go" "tests/unit/gateway/latency_tracker_test.go" false
move_test_file "internal/gateway/strategy_test.go" "tests/unit/gateway/strategy_test.go" false
move_test_file "internal/providers/openai_test.go" "tests/unit/providers/openai_test.go" false
move_test_file "internal/providers/anthropic_test.go" "tests/unit/providers/anthropic_test.go" false
move_test_file "internal/providers/local_test.go" "tests/unit/providers/local_test.go" false
move_test_file "internal/providers/google_test.go" "tests/unit/providers/google_test.go" false
move_test_file "internal/monitoring/metrics_test.go" "tests/unit/monitoring/metrics_test.go" false
move_test_file "internal/config/config_test.go" "tests/unit/config/config_test.go" false
move_test_file "internal/tunnel/domain_test.go" "tests/unit/tunnel/domain_test.go" false
move_test_file "internal/tunnel/security_test.go" "tests/unit/tunnel/security_test.go" false
move_test_file "internal/tunnel/stats_test.go" "tests/unit/tunnel/stats_test.go" false
move_test_file "internal/tunnel/auth_test.go" "tests/unit/tunnel/auth_test.go" false
move_test_file "internal/tunnel/request_tracker_test.go" "tests/unit/tunnel/request_tracker_test.go" false
move_test_file "internal/tunnel/client_test.go" "tests/unit/tunnel/client_test.go" false
move_test_file "internal/tunnel/server_test.go" "tests/unit/tunnel/server_test.go" false

# Move integration tests
move_test_file "internal/security/integration_test.go" "tests/integration/security/integration_test.go" true
move_test_file "internal/tunnel/integration_test.go" "tests/integration/tunnel/integration_test.go" true
move_test_file "internal/api/handlers/provider_keys_test.go" "tests/integration/api/handlers/provider_keys_test.go" true

echo ""
echo "Done! Review the moved files and update imports as needed."
echo "Then delete the old files once verified."

