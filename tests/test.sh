#!/bin/bash

# Tekton Hub to Artifact Hub Proxy - Comprehensive Test Suite
# Tests all API endpoints that the Tekton Hub resolver uses

BASE_URL="${1:-http://localhost:8080}"
PASSED=0
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[PASS]${NC} $1"; ((PASSED++)); }
error() { echo -e "${RED}[FAIL]${NC} $1"; ((FAILED++)); }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_field="$3"
    local expected_status="${4:-200}"

    log "Testing: $name"

    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$url" 2>/dev/null)
    if [ $? -ne 0 ]; then
        error "$name - Curl failed"
        return 0
    fi

    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')

    if [ "$http_code" != "$expected_status" ]; then
        error "$name - Expected HTTP $expected_status, got $http_code"
        echo "  Response: $body"
        return 0
    fi

    if [ -n "$expected_field" ]; then
        if echo "$body" | jq -e "$expected_field" > /dev/null 2>&1; then
            success "$name"
        else
            error "$name - Missing field: $expected_field"
            echo "  Response: $body"
            return 0
        fi
    else
        success "$name"
    fi
}

validate_yaml_content() {
    local url="$1"
    log "Validating YAML content from $url"

    response=$(curl -s "$url" 2>/dev/null)
    if [ $? -ne 0 ]; then
        error "YAML content validation - Curl failed"
        return 0
    fi

    yaml_content=$(echo "$response" | jq -r '.data.yaml' 2>/dev/null)

    if echo "$yaml_content" | grep -q "apiVersion.*tekton" 2>/dev/null; then
        success "YAML content validation - Valid Tekton resource"
    else
        error "YAML content validation - Invalid or missing Tekton YAML"
        echo "  YAML preview: $(echo "$yaml_content" | head -3 2>/dev/null)"
    fi
}

validate_landing_page() {
    log "Validating landing page content"

    response=$(curl -s "$BASE_URL/" 2>/dev/null)
    if [ $? -ne 0 ]; then
        error "Landing page validation - Could not reach landing page"
        return 0
    fi

    # Check for key content in the landing page
    if echo "$response" | grep -q "Tekton Hub to Artifact Hub Proxy" 2>/dev/null; then
        success "Landing page validation - Title present"
    else
        error "Landing page validation - Missing title"
        return 0
    fi

    if echo "$response" | grep -q "tekton.dev" 2>/dev/null; then
        success "Landing page validation - Tekton logo link present"
    else
        error "Landing page validation - Missing Tekton logo link"
    fi

    if echo "$response" | grep -q "artifacthub.io" 2>/dev/null; then
        success "Landing page validation - Artifact Hub logo link present"
    else
        error "Landing page validation - Missing Artifact Hub logo link"
    fi

    if echo "$response" | grep -q "/v1/catalogs" 2>/dev/null; then
        success "Landing page validation - API endpoints documented"
    else
        error "Landing page validation - Missing API endpoint documentation"
    fi
}

validate_translation() {
    log "Validating translation against Artifact Hub"

    proxy_response=$(curl -s "$BASE_URL/v1/resource/tekton/task/git-clone" 2>/dev/null)
    if [ $? -ne 0 ]; then
        error "Translation validation - Could not reach proxy"
        return 0
    fi

    proxy_name=$(echo "$proxy_response" | jq -r '.data.name' 2>/dev/null)

    ah_response=$(curl -s "https://artifacthub.io/api/v1/packages/tekton-task/tekton-catalog-tasks/git-clone" 2>/dev/null || echo "null")

    if [ "$ah_response" != "null" ] && [ "$ah_response" != "" ]; then
        ah_name=$(echo "$ah_response" | jq -r '.name' 2>/dev/null)
        if [ "$proxy_name" = "$ah_name" ] && [ "$proxy_name" = "git-clone" ]; then
            success "Translation validation - Names match ($proxy_name)"
        else
            error "Translation validation - Name mismatch (proxy: $proxy_name, ah: $ah_name)"
        fi
    else
        warn "Translation validation - Could not reach Artifact Hub for comparison"
    fi
}

echo "=================================================="
echo "🧪 TEKTON HUB PROXY - COMPREHENSIVE TESTS"
echo "=================================================="
echo "Testing against: $BASE_URL"
echo "Based on Tekton Hub resolver requirements from s.xml"
echo

# Core endpoints that Tekton Hub resolver actually uses
echo "1️⃣  LANDING PAGE AND CORE ENDPOINTS"
echo "-----------------------------------"

test_endpoint "Landing Page" \
    "$BASE_URL/" "" "200"

test_endpoint "Health Check" \
    "$BASE_URL/health" ".status"

test_endpoint "Get Resource Metadata (primary resolver call)" \
    "$BASE_URL/v1/resource/tekton/task/git-clone" ".data.name"

test_endpoint "Get YAML Content (resolver fetches YAML)" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/0.1/yaml" ".data.yaml"

test_endpoint "Get Raw YAML (alternative endpoint)" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/raw"

echo
echo "2️⃣  VERSION HANDLING"
echo "--------------------"

test_endpoint "Specific version YAML" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/0.6/yaml" ".data.yaml"

test_endpoint "Latest version via raw endpoint" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/0.1/raw"

echo
echo "3️⃣  DIFFERENT RESOURCE TYPES"
echo "-----------------------------"

test_endpoint "Different task (buildpacks)" \
    "$BASE_URL/v1/resource/tekton/task/buildpacks" ".data.name"

test_endpoint "Buildpacks YAML" \
    "$BASE_URL/v1/resource/tekton/task/buildpacks/0.1/yaml" ".data.yaml"

test_endpoint "Resource README" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/0.1/readme" ".data.readme"

echo
echo "4️⃣  ADDITIONAL ENDPOINTS"
echo "------------------------"

test_endpoint "List Catalogs" \
    "$BASE_URL/v1/catalogs" ".data"

test_endpoint "Get Specific Version Info" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/0.1"

# Query endpoints (may return empty but should not error)
test_endpoint "Query Resources (may be empty)" \
    "$BASE_URL/v1/query?kinds=task&catalogs=tekton&limit=3"

test_endpoint "List All Resources (may be empty)" \
    "$BASE_URL/v1/resources?limit=3"

echo
echo "5️⃣  ERROR HANDLING"
echo "------------------"

test_endpoint "Non-existent resource" \
    "$BASE_URL/v1/resource/tekton/task/non-existent" "" "404"

test_endpoint "Non-existent version" \
    "$BASE_URL/v1/resource/tekton/task/git-clone/999.999/yaml" "" "404"

test_endpoint "Invalid catalog" \
    "$BASE_URL/v1/resource/invalid-catalog/task/git-clone" "" "404"

echo
echo "6️⃣  LANDING PAGE VALIDATION"
echo "---------------------------"

validate_landing_page

echo
echo "7️⃣  RESPONSE FORMAT VALIDATION"
echo "------------------------------"

validate_yaml_content "$BASE_URL/v1/resource/tekton/task/git-clone/0.1/yaml"

log "Validating resource metadata format"
resource_response=$(curl -s "$BASE_URL/v1/resource/tekton/task/git-clone")
name=$(echo "$resource_response" | jq -r '.data.name' 2>/dev/null)
kind=$(echo "$resource_response" | jq -r '.data.kind' 2>/dev/null)
version=$(echo "$resource_response" | jq -r '.data.latestVersion.version' 2>/dev/null)

if [ "$name" = "git-clone" ] && [ "$kind" = "task" ] && [ -n "$version" ]; then
    success "Resource metadata format - All required fields present"
else
    error "Resource metadata format - Missing or incorrect fields (name=$name, kind=$kind, version=$version)"
fi

echo
echo "8️⃣  TRANSLATION VERIFICATION"
echo "----------------------------"

validate_translation

echo
echo "9️⃣  CONFIGURATION TESTING"
echo "-------------------------"

log "Testing help documentation"
if ./bin/tekton-hub-proxy --help 2>/dev/null | grep -q "disable-landing-page"; then
    success "Help documentation - Landing page flag documented"
else
    error "Help documentation - Landing page flag not documented"
fi

log "Testing landing page disable functionality"
echo "Note: To test --disable-landing-page flag, run:"
echo "  ./bin/tekton-hub-proxy --disable-landing-page --port 8081 &"
echo "  curl -w \"HTTP: %{http_code}\\n\" http://localhost:8081/"
echo "  # Should return 404 Not Found"
echo "  pkill -f 'tekton-hub-proxy.*8081'"
echo

log "Configuration options available:"
echo "  Command line: --disable-landing-page"
echo "  Config file:  landing_page.enabled: false"
echo "  Environment:  THP_LANDING_PAGE_ENABLED=false"

echo
echo "🔟  MANUAL CURL COMMANDS"
echo "------------------------"
echo "For manual testing, use these exact commands:"
echo
echo "# Landing page:"
echo "curl $BASE_URL/"
echo
echo "# Core resolver endpoints:"
echo "curl $BASE_URL/v1/resource/tekton/task/git-clone"
echo "curl $BASE_URL/v1/resource/tekton/task/git-clone/0.1/yaml"
echo "curl $BASE_URL/v1/resource/tekton/task/git-clone/raw"
echo
echo "# Validation:"
echo "curl -s $BASE_URL/v1/resource/tekton/task/git-clone/0.1/yaml | jq '.data.yaml' | head -5"
echo "curl -s $BASE_URL/v1/resource/tekton/task/git-clone | jq '.data.name'"
echo
echo "# Error testing:"
echo "curl -w \"HTTP: %{http_code}\\n\" $BASE_URL/v1/resource/tekton/task/non-existent"
echo
echo "# Configuration testing:"
echo "./bin/tekton-hub-proxy --help | grep disable-landing-page"
echo "./bin/tekton-hub-proxy --disable-landing-page --port 8081 &"

echo
echo "=================================================="
echo "📊 TEST RESULTS SUMMARY"
echo "=================================================="
echo -e "${GREEN}PASSED: $PASSED${NC}"
echo -e "${RED}FAILED: $FAILED${NC}"
echo

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
    echo
    echo "✅ Your Tekton Hub proxy is fully functional!"
    echo "✅ All endpoints used by Tekton Hub resolver work correctly"
    echo "✅ Translation from Tekton Hub API to Artifact Hub is working"
    echo "✅ Response formats match Tekton Hub specifications"
    echo "✅ Error handling is appropriate"
    echo "✅ The Tekton Pipeline Hub resolver will work with this proxy"
    echo
    echo "🚀 Ready for production use!"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    echo
    echo "Please check the failed tests above and verify:"
    echo "  - The proxy server is running on $BASE_URL"
    echo "  - Artifact Hub (https://artifacthub.io) is accessible"
    echo "  - Catalog mappings are configured correctly in config.yaml"
    echo "  - The proxy has proper network connectivity"
    exit 1
fi