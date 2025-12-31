#!/bin/bash

# Script to add a new application to the NKP catalog repository
# Usage: ./add-application.sh --appname <name> --version <version> --ocirepo <oci-repo-path>
# Example: ./add-application.sh --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize variables
APPNAME=""
VERSION=""
OCIREPO=""
FORCE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --appname)
            APPNAME="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --ocirepo)
            OCIREPO="$2"
            shift 2
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 --appname <name> --version <version> --ocirepo <oci-repo-path>"
            echo ""
            echo "Options:"
            echo "  --appname    Application name (e.g., podinfo, vault)"
            echo "  --version    Application version (e.g., 6.9.4, 0.31.0)"
            echo "  --ocirepo    OCI repository path (e.g., oci://ghcr.io/stefanprodan/charts/podinfo)"
            echo "  --force      Skip confirmation prompt if application directory exists"
            echo ""
            echo "Example:"
            echo "  $0 --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo"
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$APPNAME" ]; then
    echo -e "${RED}Error: --appname is required${NC}"
    echo "Use --help for usage information"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: --version is required${NC}"
    echo "Use --help for usage information"
    exit 1
fi

if [ -z "$OCIREPO" ]; then
    echo -e "${RED}Error: --ocirepo is required${NC}"
    echo "Use --help for usage information"
    exit 1
fi

# Validate OCI repo format
if [[ ! "$OCIREPO" =~ ^oci:// ]]; then
    echo -e "${YELLOW}Warning: OCI repo path should start with 'oci://'. Adding it automatically...${NC}"
    OCIREPO="oci://${OCIREPO}"
fi

# Get the repository root directory
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${REPO_DIR}/applications/${APPNAME}/${VERSION}"
HELMRELEASE_FILE="${APP_DIR}/helmrelease/helmrelease.yaml"
METADATA_FILE="${APP_DIR}/metadata.yaml"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Adding Application to NKP Catalog${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Application: ${GREEN}${APPNAME}${NC}"
echo -e "Version: ${GREEN}${VERSION}${NC}"
echo -e "OCI Repository: ${GREEN}${OCIREPO}${NC}"
echo ""

# Step 1: Generate catalog repository structure
echo -e "${YELLOW}Step 1: Generating catalog repository structure...${NC}"
if [ -d "$APP_DIR" ]; then
    echo -e "${YELLOW}Warning: Application directory already exists: ${APP_DIR}${NC}"
    if [ "$FORCE" = false ]; then
        read -p "Do you want to continue? This may overwrite existing files. (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Aborted by user${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}  --force flag set, proceeding without confirmation${NC}"
    fi
fi

nkp generate catalog-repository --apps="${APPNAME}=${VERSION}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to generate catalog repository structure!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Catalog repository structure generated${NC}"
echo ""

# Step 2: Update OCI repository path in helmrelease.yaml
echo -e "${YELLOW}Step 2: Updating OCI repository path in helmrelease.yaml...${NC}"
if [ ! -f "$HELMRELEASE_FILE" ]; then
    echo -e "${RED}Error: helmrelease.yaml not found at ${HELMRELEASE_FILE}${NC}"
    exit 1
fi

# Use sed to replace the OCI repository URL
# This looks for the url: field in the OCIRepository resource and replaces it
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    sed -i '' "s|url: oci://.*|url: ${OCIREPO}|" "$HELMRELEASE_FILE"
else
    # Linux uses GNU sed
    sed -i "s|url: oci://.*|url: ${OCIREPO}|" "$HELMRELEASE_FILE"
fi

echo -e "${GREEN}✓ OCI repository path updated${NC}"
echo ""

# Step 3: Update metadata.yaml with basic information
echo -e "${YELLOW}Step 3: Updating metadata.yaml...${NC}"
if [ ! -f "$METADATA_FILE" ]; then
    echo -e "${RED}Error: metadata.yaml not found at ${METADATA_FILE}${NC}"
    exit 1
fi

# Create a capitalized display name from appname
# Convert kebab-case to Title Case
DISPLAY_NAME=$(echo "$APPNAME" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2)); print}')

# Update displayName
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    sed -i '' "s/^displayName:.*/displayName: ${DISPLAY_NAME}/" "$METADATA_FILE"
else
    # Linux uses GNU sed
    sed -i "s/^displayName:.*/displayName: ${DISPLAY_NAME}/" "$METADATA_FILE"
fi

# Check if supportLink is empty and set a default GitHub URL
CURRENT_SUPPORT_LINK=$(grep "^supportLink:" "$METADATA_FILE" | sed 's/supportLink: *//' | sed 's/"//g')
if [[ -z "$CURRENT_SUPPORT_LINK" ]] || [[ "$CURRENT_SUPPORT_LINK" == '""' ]] || [[ "$CURRENT_SUPPORT_LINK" == "''" ]]; then
    # Set a default GitHub URL based on appname (common pattern)
    DEFAULT_SUPPORT_LINK="https://github.com/${APPNAME}/${APPNAME}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|^supportLink:.*|supportLink: ${DEFAULT_SUPPORT_LINK}|" "$METADATA_FILE"
    else
        sed -i "s|^supportLink:.*|supportLink: ${DEFAULT_SUPPORT_LINK}|" "$METADATA_FILE"
    fi
    echo -e "${YELLOW}  Set default supportLink to: ${DEFAULT_SUPPORT_LINK}${NC}"
    echo -e "${YELLOW}  Note: You may want to update this with the correct support URL${NC}"
fi

# Note: The description field in metadata.yaml is typically a multi-line YAML field
# We'll leave it as-is since nkp generate should create a reasonable default
# Users can manually update it with more detailed information

echo -e "${GREEN}✓ Metadata.yaml updated with basic information${NC}"
echo -e "${YELLOW}  Note: You may want to manually update metadata.yaml with more detailed information${NC}"
echo ""

# Step 4: Validate the catalog repository
echo -e "${YELLOW}Step 4: Validating catalog repository...${NC}"
nkp validate catalog-repository --repo-dir="${REPO_DIR}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Validation failed!${NC}"
    echo -e "${YELLOW}Please check the errors above and fix them manually${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Validation passed${NC}"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Application added successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Application directory: ${GREEN}${APP_DIR}${NC}"
echo -e "HelmRelease file: ${GREEN}${HELMRELEASE_FILE}${NC}"
echo -e "Metadata file: ${GREEN}${METADATA_FILE}${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review and update ${METADATA_FILE} with detailed information"
echo "  2. Review ${HELMRELEASE_FILE} to ensure OCI repository path is correct"
echo "  3. Update any configuration in ${APP_DIR}/helmrelease/cm.yaml if needed"
echo "  4. Run validation again: nkp validate catalog-repository --repo-dir=."
echo "  5. Build and push the catalog: ./build-and-push.sh <tag>"
echo ""

