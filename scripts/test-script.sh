#!/bin/bash

# Enhanced test script that uses the spacectl CLI to test the API
# Provides detailed output, error handling, and comprehensive summary

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Generate random 3-letter suffix to avoid conflicts
RANDOM_SUFFIX=$(openssl rand -hex 3 | cut -c1-3 | tr '[:lower:]' '[:upper:]')

# Test resource names with random suffix
TEST_ORG_NAME="test-${RANDOM_SUFFIX}"
TEST_PROJECT_NAME="test-${RANDOM_SUFFIX}"
TEST_TENANT_NAME="test-${RANDOM_SUFFIX}"

# Test tracking variables
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_COMMANDS=()
DEBUG_MODE=false
TEST_PROJECT_ID=""
TENANT_CREATION_SUCCESS=false

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    local timestamp=$(date '+%H:%M:%S')
    
    case $status in
        "INFO")
            echo -e "${BLUE}[$timestamp] INFO:${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[$timestamp] ✓ SUCCESS:${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[$timestamp] ✗ ERROR:${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[$timestamp] ⚠ WARNING:${NC} $message"
            ;;
    esac
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --debug)
                DEBUG_MODE=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --debug    Enable debug mode (shows full command output)"
    echo "  --help     Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  TEST_EMAIL     Email for authentication"
    echo "  TEST_PASSWORD   Password for authentication"
}

# Function to capture project ID from project list
capture_project_id() {
    local project_name="$1"
    local org_name="$2"
    local project_id

    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "INFO" "Capturing project ID for project: $project_name in organization: $org_name"
        # Show what projects are available
        print_status "INFO" "All available projects in organization '$org_name':"
        ./bin/spacectl project list --org-name "$org_name" --output json 2>/dev/null | jq -r '.[] | "\(.name) (\(.id))"' 2>/dev/null || echo "Could not parse project list"
    fi

    # Find project by listing all projects in the organization and matching by name

    local project_list_output
    project_list_output=$(./bin/spacectl project list --org-name "$org_name" --output json 2>/dev/null)
    local list_exit_code=$?

    if [ $list_exit_code -eq 0 ] && [ -n "$project_list_output" ]; then
        if command -v jq >/dev/null 2>&1; then
            project_id=$(echo "$project_list_output" | jq -r ".[] | select(.name == \"$project_name\") | .id" 2>/dev/null)
            if [ -n "$project_id" ] && [ "$project_id" != "null" ]; then
                TEST_PROJECT_ID="$project_id"
                if [ "$DEBUG_MODE" = "true" ]; then
                    print_status "SUCCESS" "Captured project ID: $TEST_PROJECT_ID"
                fi
                return 0
            fi
        fi
    fi

    print_status "ERROR" "Failed to capture project ID for: $project_name"
    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "ERROR" "Project list output:"
        echo "$project_list_output"
        print_status "ERROR" "Available projects in organization '$org_name':"
        ./bin/spacectl project list --org-name "$org_name" --output json 2>/dev/null | jq -r '.[] | "\(.name) (\(.id))"' 2>/dev/null || echo "Could not parse project list"
    fi
    return 1
}

# Function to run a test command and track results
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_success="${3:-true}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_status "INFO" "Running test: $test_name"
    
    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "INFO" "Command: $command"
    fi
    
    local start_time=$(date +%s)
    
    # Capture output for debug mode
    local output
    local exit_code
    
    if [ "$DEBUG_MODE" = "true" ]; then
        output=$(eval "$command" 2>&1)
        exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            print_status "SUCCESS" "Command output:"
            echo "$output"
        else
            print_status "ERROR" "Command failed with exit code $exit_code"
            print_status "ERROR" "Command output:"
            echo "$output"
        fi
    else
        output=$(eval "$command" 2>&1)
        exit_code=$?
    fi
    
    if [ $exit_code -eq 0 ]; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        if [ "$expected_success" = "true" ]; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
            print_status "SUCCESS" "$test_name completed successfully (${duration}s)"
            # Track tenant creation success
            if [ "$test_name" = "Create Test Tenant" ]; then
                TENANT_CREATION_SUCCESS=true
            fi
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_COMMANDS+=("$test_name: Expected failure but command succeeded")
            print_status "ERROR" "$test_name should have failed but succeeded (${duration}s)"
        fi
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        if [ "$expected_success" = "false" ]; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
            print_status "SUCCESS" "$test_name failed as expected (${duration}s)"
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_COMMANDS+=("$test_name: Command failed with exit code $exit_code")
            print_status "ERROR" "$test_name failed unexpectedly (${duration}s)"
        fi
    fi
    
    echo ""
}

# Function to check if tenant exists by trying to get it directly
check_tenant_exists() {
    local tenant_name="$1"
    local project_id="$2"

    # Try to get the tenant directly using the project ID
    local cmd="./bin/spacectl tenant get --name \"$tenant_name\" --project \"$project_id\""

    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "INFO" "Running command: $cmd"
    fi

    local output
    output=$(./bin/spacectl tenant get --name "$tenant_name" --project "$project_id" 2>&1)
    local exit_code=$?

    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "INFO" "Direct tenant check for '$tenant_name' in project '$project_id' (exit code: $exit_code)"
        if [ $exit_code -eq 0 ]; then
            print_status "SUCCESS" "Tenant found via direct get:"
            echo "$output"
        else
            print_status "INFO" "Tenant not found via direct get. Error output:"
            echo "$output"
        fi
    fi

    # If command succeeded, check if status is "ready" or "failed"
    if [ $exit_code -eq 0 ]; then
        # Extract status from output (last column)
        local status=$(echo "$output" | grep -v "^STATUS" | grep -v "^----" | awk '{print $NF}' | grep -v "^$" | tail -1)

        if [ "$DEBUG_MODE" = "true" ]; then
            print_status "INFO" "Tenant status: '$status'"
        fi

        # Check for failed status
        if [ "$status" = "failed" ]; then
            return 2  # Special return code for failed status
        fi

        # Only return success if status is "ready"
        if [ "$status" = "ready" ]; then
            return 0
        else
            # Tenant exists but not ready yet
            return 1
        fi
    fi

    return $exit_code
}

# Function to wait for tenant creation
wait_for_tenant() {
    local tenant_name="$1"
    local max_wait_time="${2:-180}"  # Default 180 seconds (3 minutes)
    local wait_interval="${3:-5}"  # Default 5 seconds between checks
    
    print_status "INFO" "Waiting for tenant '$tenant_name' to be created (max ${max_wait_time}s)..."
    
    local start_time=$(date +%s)
    local end_time=$((start_time + max_wait_time))
    local check_count=0
    
    while [ $(date +%s) -lt $end_time ]; do
        check_count=$((check_count + 1))
        
        # Verify project ID is correct before checking tenant
        if [ "$DEBUG_MODE" = "true" ] && [ -n "$TEST_PROJECT_ID" ]; then
            print_status "INFO" "Verifying project ID '$TEST_PROJECT_ID' exists..."
            project_check=$(./bin/spacectl project get --project-id "$TEST_PROJECT_ID" --output json 2>/dev/null)
            project_exit_code=$?
            if [ $project_exit_code -eq 0 ]; then
                project_name=$(echo "$project_check" | jq -r '.name' 2>/dev/null)
                print_status "SUCCESS" "Project ID '$TEST_PROJECT_ID' verified - belongs to project '$project_name'"
            else
                print_status "ERROR" "Project ID '$TEST_PROJECT_ID' not found or invalid"
                print_status "ERROR" "Available projects (all organizations):"
                ./bin/spacectl project list --all --output json 2>/dev/null | jq -r '.[] | "\(.name) (\(.id))"' 2>/dev/null || echo "Could not parse project list"
            fi
        fi
        
        # First try direct tenant get (more reliable)
        if [ -n "$TEST_PROJECT_ID" ]; then
            check_tenant_exists "$tenant_name" "$TEST_PROJECT_ID"
            local tenant_check_result=$?

            if [ $tenant_check_result -eq 0 ]; then
                local current_time=$(date +%s)
                local duration=$((current_time - start_time))
                print_status "SUCCESS" "Tenant '$tenant_name' found via direct get after ${duration}s"
                return 0
            elif [ $tenant_check_result -eq 2 ]; then
                local current_time=$(date +%s)
                local duration=$((current_time - start_time))
                print_status "ERROR" "Tenant '$tenant_name' is in failed state after ${duration}s"
                return 2
            fi
        fi
        
        # Fallback to tenant list method
        if [ "$DEBUG_MODE" = "true" ]; then
            if [ -n "$TEST_PROJECT_ID" ]; then
                print_status "INFO" "Running command: ./bin/spacectl tenant list --project \"$TEST_PROJECT_ID\""
            else
                print_status "INFO" "Running command: ./bin/spacectl tenant list"
            fi
        fi

        local tenant_list_output
        if [ -n "$TEST_PROJECT_ID" ]; then
            tenant_list_output=$(./bin/spacectl tenant list --project "$TEST_PROJECT_ID" 2>&1)
        else
            tenant_list_output=$(./bin/spacectl tenant list 2>&1)
        fi
        local list_exit_code=$?

        if [ "$DEBUG_MODE" = "true" ]; then
            print_status "INFO" "Tenant list check #${check_count} (exit code: $list_exit_code):"
            echo "$tenant_list_output"
        fi
        
        if [ $list_exit_code -eq 0 ] && [ -n "$tenant_list_output" ]; then
            # Check if tenant name appears in the list and extract its status
            local tenant_line=$(echo "$tenant_list_output" | grep "^$tenant_name")

            if [ -n "$tenant_line" ]; then
                # Extract status from the line (last column)
                local status=$(echo "$tenant_line" | awk '{print $NF}')

                if [ "$DEBUG_MODE" = "true" ]; then
                    print_status "INFO" "Tenant found in list with status: '$status'"
                fi

                # Check for failed status
                if [ "$status" = "failed" ]; then
                    return 2  # Special return code for failed status
                fi

                # Only return success if status is "ready"
                if [ "$status" = "ready" ]; then
                    local current_time=$(date +%s)
                    local duration=$((current_time - start_time))
                    print_status "SUCCESS" "Tenant '$tenant_name' is ready after ${duration}s"
                    return 0
                else
                    if [ "$DEBUG_MODE" = "true" ]; then
                        print_status "INFO" "Tenant exists but status is '$status', waiting for 'ready'..."
                    fi
                fi
            fi

            # Also check for partial matches or similar names in debug mode
            if [ "$DEBUG_MODE" = "true" ]; then
                print_status "INFO" "Available tenants:"
                echo "$tenant_list_output" | grep -v "^NAME" | grep -v "^----" | grep -v "^$" | awk '{print $1}' | head -10
            fi
        else
            if [ "$DEBUG_MODE" = "true" ]; then
                print_status "WARNING" "Tenant list command failed or returned empty output"
            fi
        fi
        
        print_status "INFO" "Tenant '$tenant_name' not ready yet, waiting ${wait_interval}s..."
        sleep $wait_interval
    done
    
    print_status "ERROR" "Timeout waiting for tenant '$tenant_name' to be created"
    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "ERROR" "Final tenant list output:"
        ./bin/spacectl tenant list --quiet 2>&1 || echo "Failed to get tenant list"
        if [ -n "$TEST_PROJECT_ID" ]; then
            print_status "ERROR" "Final direct tenant check:"
            ./bin/spacectl tenant get --name "$tenant_name" --project "$TEST_PROJECT_ID" 2>&1 || echo "Failed to get tenant directly"
        fi
    fi
    return 1
}

# Function to validate environment
validate_environment() {
    print_status "INFO" "Validating environment variables..."
    
    if [ -z "$TEST_EMAIL" ]; then
        print_status "ERROR" "TEST_EMAIL environment variable is not set"
        exit 1
    fi
    
    if [ -z "$TEST_PASSWORD" ]; then
        print_status "ERROR" "TEST_PASSWORD environment variable is not set"
        exit 1
    fi
    
    print_status "SUCCESS" "Environment variables validated"
    echo ""
}

# Function to print summary
print_summary() {
    echo "=========================================="
    echo -e "${BLUE}TEST EXECUTION SUMMARY${NC}"
    echo "=========================================="
    echo ""
    
    echo -e "Total Tests Run: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
        echo ""
        echo "The spacectl CLI is working correctly. All operations completed successfully:"
        echo "• Authentication and login"
        echo "• Organization management (create, list, delete)"
        echo "• Project management (create, list, delete)"
        echo "• Tenant management (create, list, delete with proper waiting)"
        echo "• Cloud provider and region information retrieval"
        echo "• Kubernetes version information retrieval"
        echo "• Cleanup operations completed without errors"
    else
        echo -e "${RED}❌ SOME TESTS FAILED ❌${NC}"
        echo ""
        echo "The following tests encountered issues:"
        for failed_cmd in "${FAILED_COMMANDS[@]}"; do
            echo -e "  ${RED}• $failed_cmd${NC}"
        done
        echo ""
        echo "Please check the error messages above and ensure:"
        echo "• The API server is running and accessible"
        echo "• Authentication credentials are correct"
        echo "• Network connectivity is available"
        echo "• Required permissions are granted"
    fi
    
    echo ""
    echo "=========================================="
    
    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Main execution
echo -e "${BLUE}Starting spacectl CLI Test Suite${NC}"
echo "=========================================="
echo ""

# Parse command line arguments
parse_args "$@"

# Show generated test resource names
print_status "INFO" "Generated test resource names:"
print_status "INFO" "  Organization: $TEST_ORG_NAME"
print_status "INFO" "  Project: $TEST_PROJECT_NAME"
print_status "INFO" "  Tenant: $TEST_TENANT_NAME"
echo ""

# Validate environment before starting
validate_environment

# Build the CLI
print_status "INFO" "Building spacectl CLI..."
if make build > /dev/null 2>&1; then
    print_status "SUCCESS" "spacectl CLI built successfully"
else
    print_status "ERROR" "Failed to build spacectl CLI"
    exit 1
fi
echo ""

# Show version information
run_test "Show CLI Version" "./bin/spacectl version"

# Run test suite
run_test "Authentication Login" "./bin/spacectl auth login --email $TEST_EMAIL --password $TEST_PASSWORD"
run_test "List Organizations" "./bin/spacectl org list"
run_test "Create Test Organization" "./bin/spacectl org create $TEST_ORG_NAME --description \"Test organization\""
run_test "List Projects" "./bin/spacectl project list --org-name $TEST_ORG_NAME"
run_test "Create Test Project" "./bin/spacectl project create $TEST_PROJECT_NAME --description \"Test project\" --org-name $TEST_ORG_NAME"

# Small delay to ensure project is fully available
print_status "INFO" "Waiting 2 seconds for project to be fully available..."
sleep 2

# Show all available projects after creation for debugging
if [ "$DEBUG_MODE" = "true" ]; then
    print_status "INFO" "Available projects in organization '$TEST_ORG_NAME' after creation:"
    ./bin/spacectl project list --org-name "$TEST_ORG_NAME" --output json 2>/dev/null | jq -r '.[] | "\(.name) (\(.id))"' 2>/dev/null || echo "Could not parse project list"
fi

# Capture project ID for tenant creation
if capture_project_id "$TEST_PROJECT_NAME" "$TEST_ORG_NAME"; then
    print_status "SUCCESS" "Project ID captured: $TEST_PROJECT_ID"
    
    # Verify the project ID is correct by checking if it matches the project we just created
    if [ "$DEBUG_MODE" = "true" ]; then
        print_status "INFO" "Verifying captured project ID matches the created project..."
        project_details=$(./bin/spacectl project get --project-id "$TEST_PROJECT_ID" --output json 2>/dev/null)
        if [ $? -eq 0 ]; then
            project_name=$(echo "$project_details" | jq -r '.name' 2>/dev/null)
            if [ "$project_name" = "$TEST_PROJECT_NAME" ]; then
                print_status "SUCCESS" "Project ID verification passed - '$TEST_PROJECT_ID' belongs to project '$TEST_PROJECT_NAME'"
            else
                print_status "ERROR" "Project ID mismatch - '$TEST_PROJECT_ID' belongs to project '$project_name', not '$TEST_PROJECT_NAME'"
                print_status "ERROR" "This will cause tenant operations to fail"
            fi
        else
            print_status "ERROR" "Could not verify project ID - project not found"
        fi
    fi
    
    run_test "List Tenants" "./bin/spacectl tenant list --quiet"
    run_test "List Available Locations" "./bin/spacectl tenant locations"
    run_test "List Available Kubernetes Versions" "./bin/spacectl tenant k8s-versions"
    run_test "Create Test Tenant" "./bin/spacectl tenant create $TEST_TENANT_NAME --project $TEST_PROJECT_ID --cloud eks --region eu --k8s-version v1.33.1 --compute 2 --memory 4"
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_COMMANDS+=("Project ID Capture: Failed to capture project ID")
    print_status "ERROR" "Skipping tenant tests due to project ID capture failure"
fi

# Wait for tenant creation before proceeding (only if tenant creation was successful)
if [ "$TENANT_CREATION_SUCCESS" = "true" ]; then
    wait_for_tenant "$TEST_TENANT_NAME" 180 5
    WAIT_RESULT=$?

    if [ $WAIT_RESULT -eq 0 ]; then
        # Tenant is ready - proceed with kubeconfig and nginx tests
        print_status "INFO" "Retrieving tenant ID for kubeconfig download..."
        TENANT_ID=$(./bin/spacectl tenant get --name "$TEST_TENANT_NAME" --project "$TEST_PROJECT_ID" --output json 2>/dev/null | jq -r '.id' 2>/dev/null)

        if [ -n "$TENANT_ID" ] && [ "$TENANT_ID" != "null" ]; then
            print_status "SUCCESS" "Tenant ID: $TENANT_ID"

            # Download kubeconfig
            KUBECONFIG_FILE="/tmp/test-kubeconfig-${TENANT_ID}.yaml"
            print_status "INFO" "Downloading kubeconfig to $KUBECONFIG_FILE..."
            if ./bin/spacectl tenant kubeconfig "$TENANT_ID" --output-file "$KUBECONFIG_FILE" 2>&1; then
                print_status "SUCCESS" "Kubeconfig downloaded successfully"

                # Deploy nginx pod with minimal resource limits
                print_status "INFO" "Deploying nginx pod to test cluster connectivity..."
                if [ "$DEBUG_MODE" = "true" ]; then
                    print_status "INFO" "Running: kubectl run nginx-test with resource limits (cpu: 100m-200m, memory: 128Mi-256Mi)"
                fi

                POD_CREATE_OUTPUT=$(kubectl --kubeconfig="$KUBECONFIG_FILE" run nginx-test --image=mirror.gcr.io/library/nginx:latest --restart=Never --overrides='{
                  "spec": {
                    "containers": [{
                      "name": "nginx-test",
                      "image": "mirror.gcr.io/library/nginx:latest",
                      "resources": {
                        "requests": {
                          "cpu": "100m",
                          "memory": "128Mi"
                        },
                        "limits": {
                          "cpu": "200m",
                          "memory": "256Mi"
                        }
                      }
                    }]
                  }
                }' 2>&1)
                if [ $? -eq 0 ]; then
                    print_status "SUCCESS" "Nginx pod created"
                    if [ "$DEBUG_MODE" = "true" ]; then
                        echo "$POD_CREATE_OUTPUT"
                    fi

                    # Wait for pod to be ready
                    print_status "INFO" "Waiting for nginx pod to be ready (max 60s)..."
                    if [ "$DEBUG_MODE" = "true" ]; then
                        print_status "INFO" "Running: kubectl --kubeconfig=\"$KUBECONFIG_FILE\" wait --for=condition=Ready pod/nginx-test --timeout=60s"
                    fi

                    WAIT_OUTPUT=$(kubectl --kubeconfig="$KUBECONFIG_FILE" wait --for=condition=Ready pod/nginx-test --timeout=60s 2>&1)
                    if [ $? -eq 0 ]; then
                        print_status "SUCCESS" "Nginx pod is ready"

                        # Verify pod is running
                        print_status "INFO" "Verifying pod status..."
                        POD_STATUS=$(kubectl --kubeconfig="$KUBECONFIG_FILE" get pod nginx-test -o jsonpath='{.status.phase}' 2>/dev/null)
                        if [ "$POD_STATUS" = "Running" ]; then
                            print_status "SUCCESS" "Pod verification passed - nginx is running"
                            PASSED_TESTS=$((PASSED_TESTS + 1))
                        else
                            print_status "WARNING" "Pod status is '$POD_STATUS' instead of 'Running'"
                        fi

                        # Clean up nginx pod
                        print_status "INFO" "Cleaning up nginx pod..."
                        if [ "$DEBUG_MODE" = "true" ]; then
                            print_status "INFO" "Running: kubectl --kubeconfig=\"$KUBECONFIG_FILE\" delete pod nginx-test"
                        fi
                        DELETE_OUTPUT=$(kubectl --kubeconfig="$KUBECONFIG_FILE" delete pod nginx-test --wait=true --timeout=30s 2>&1)
                        if [ $? -eq 0 ]; then
                            print_status "SUCCESS" "Nginx pod deleted"
                        else
                            print_status "WARNING" "Failed to delete nginx pod, continuing anyway"
                            if [ "$DEBUG_MODE" = "true" ]; then
                                echo "$DELETE_OUTPUT"
                            fi
                        fi
                    else
                        print_status "ERROR" "Nginx pod failed to become ready"
                        echo "$WAIT_OUTPUT"

                        # Show pod status for debugging
                        print_status "INFO" "Pod status details:"
                        kubectl --kubeconfig="$KUBECONFIG_FILE" get pod nginx-test -o wide 2>&1 || echo "Could not get pod status"

                        print_status "INFO" "Pod events:"
                        kubectl --kubeconfig="$KUBECONFIG_FILE" describe pod nginx-test 2>&1 | grep -A 20 "Events:" || echo "Could not get pod events"

                        print_status "INFO" "Pod logs:"
                        kubectl --kubeconfig="$KUBECONFIG_FILE" logs nginx-test 2>&1 || echo "Could not get pod logs"

                        FAILED_TESTS=$((FAILED_TESTS + 1))

                        # Clean up failed pod
                        print_status "INFO" "Cleaning up failed pod..."
                        kubectl --kubeconfig="$KUBECONFIG_FILE" delete pod nginx-test --force --grace-period=0 2>&1 || echo "Could not delete pod"
                    fi
                else
                    print_status "ERROR" "Failed to create nginx pod"
                    echo "$POD_CREATE_OUTPUT"
                    FAILED_TESTS=$((FAILED_TESTS + 1))
                fi

                # Clean up kubeconfig file
                rm -f "$KUBECONFIG_FILE"
            else
                print_status "ERROR" "Failed to download kubeconfig"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            print_status "ERROR" "Failed to retrieve tenant ID"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi

        # Delete tenant
        run_test "Delete Test Tenant" "./bin/spacectl tenant delete --name $TEST_TENANT_NAME --project $TEST_PROJECT_ID --force"
    elif [ $WAIT_RESULT -eq 2 ]; then
        # Tenant is in failed state - skip tests and proceed to deletion
        print_status "WARNING" "Tenant is in failed state, skipping kubeconfig and nginx tests"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_COMMANDS+=("Tenant Deployment: Tenant entered failed state")

        # Still delete the failed tenant to clean up
        run_test "Delete Failed Tenant" "./bin/spacectl tenant delete --name $TEST_TENANT_NAME --project $TEST_PROJECT_ID --force"
    else
        # Timeout waiting for tenant
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_COMMANDS+=("Tenant Creation Wait: Timeout waiting for tenant to be created")
        print_status "ERROR" "Timeout waiting for tenant, attempting deletion anyway..."

        # Try to delete the tenant even if it timed out
        run_test "Delete Timed Out Tenant" "./bin/spacectl tenant delete --name $TEST_TENANT_NAME --project $TEST_PROJECT_ID --force"
    fi
else
    print_status "ERROR" "Skipping tenant wait and deletion due to tenant creation failure"
fi

run_test "Delete Test Project" "./bin/spacectl project delete --name $TEST_PROJECT_NAME --force"
run_test "Delete Test Organization" "./bin/spacectl org delete --name $TEST_ORG_NAME --force"

# Print final summary
print_summary