#!/bin/bash

# Check for latest available chart versions for catalog apps and recommend add-app commands.
# Usage: ./check-latest-versions.sh [--appname <name>] [--all]
# Requires: curl, and for Helm-repo apps: helm (with repo added or --helmrepo-url in mapping).

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APPLICATIONS_DIR="${REPO_DIR}/applications"

# Per-app meta file: for apps that use Helm repo + ocipush, store source so we check the Helm repo for new versions.
# See applications/<app>/.catalog-source.yaml and docs/CATALOG-SOURCE.md
CATALOG_SOURCE_FILE=".catalog-source.yaml"

# Read Helm-repo source from applications/<app>/.catalog-source.yaml (helmrepo, helmrepoUrl, ocipush).
# Output: three lines (helmrepo, helmrepoUrl, ocipush). Empty line if key missing. Returns 0 if file exists and helmrepo is set.
get_helmrepo_source() {
    local app="$1"
    local f="${APPLICATIONS_DIR}/${app}/${CATALOG_SOURCE_FILE}"
    [[ ! -f "$f" ]] && return 1
    local helmrepo helmrepo_url ocipush
    helmrepo=$(grep -E '^helmrepo:' "$f" 2>/dev/null | sed -E 's/^helmrepo:[[:space:]]*//;s/[[:space:]]*$//' | head -1)
    [[ -z "$helmrepo" ]] && return 1
    helmrepo_url=$(grep -E '^helmrepoUrl:' "$f" 2>/dev/null | sed -E 's/^helmrepoUrl:[[:space:]]*//;s/[[:space:]]*$//' | head -1)
    ocipush=$(grep -E '^ocipush:' "$f" 2>/dev/null | sed -E 's/^ocipush:[[:space:]]*//;s/[[:space:]]*$//' | head -1)
    echo "$helmrepo"
    echo "$helmrepo_url"
    echo "$ocipush"
    return 0
}

# Get OCI URL from app's helmrelease (from latest version dir).
get_oci_url_for_app() {
    local app="$1"
    local app_dir="${APPLICATIONS_DIR}/${app}"
    [[ ! -d "$app_dir" ]] && return 1
    local latest_ver
    latest_ver=$(ls -1 "$app_dir" 2>/dev/null | sort -V | tail -1)
    [[ -z "$latest_ver" ]] && return 1
    local hr="${app_dir}/${latest_ver}/helmrelease/helmrelease.yaml"
    [[ ! -f "$hr" ]] && return 1
    # OCIRepository url: may be "url: oci://..." or "url: oci://..."
    local url
    url=$(grep -A 200 'kind: OCIRepository' "$hr" | grep '^\s*url:' | head -1 | sed -E 's/.*url:\s*//' | tr -d ' "\047')
    [[ -n "$url" ]] && echo "$url" && return 0
    return 1
}

# Get all version dirs for an app (sorted).
get_catalog_versions() {
    local app="$1"
    local app_dir="${APPLICATIONS_DIR}/${app}"
    [[ ! -d "$app_dir" ]] && return
    ls -1 "$app_dir" 2>/dev/null | sort -V
}

# Get latest catalog version for an app.
get_latest_catalog_version() {
    get_catalog_versions "$1" | tail -1
}

# List tags from OCI registry (Docker/OCI Distribution API v2).
# Supports: ghcr.io, quay.io, registry-1.docker.io, registry.k8s.io
oci_list_tags() {
    local oci_url="$1"
    # oci_url is like oci://ghcr.io/org/repo or oci://quay.io/org/repo
    local url_no_oci="${oci_url#oci://}"
    local host path
    host="${url_no_oci%%/*}"
    path="${url_no_oci#*/}"

    local tags_list_url
    local auth_header=""

    case "$host" in
        ghcr.io)
            # GHCR: token required for tags/list
            local scope="repository:${path}:pull"
            local token
            token=$(curl -sL --connect-timeout 10 --max-time 30 "https://ghcr.io/token?scope=${scope}" 2>/dev/null | sed -n 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' || true)
            if [[ -n "$token" ]]; then
                auth_header="Authorization: Bearer $token"
            fi
            tags_list_url="https://${host}/v2/${path}/tags/list"
            ;;
        quay.io)
            tags_list_url="https://quay.io/v2/${path}/tags/list"
            ;;
        registry.k8s.io)
            tags_list_url="https://${host}/v2/${path}/tags/list"
            ;;
        registry-1.docker.io)
            # Docker Hub: path might be library/foo or org/foo
            tags_list_url="https://registry-1.docker.io/v2/${path}/tags/list"
            ;;
        *)
            # Generic v2 API
            tags_list_url="https://${host}/v2/${path}/tags/list"
            ;;
    esac

    local json
    if [[ -n "$auth_header" ]]; then
        json=$(curl -sL --connect-timeout 10 --max-time 30 -H "$auth_header" "$tags_list_url" 2>/dev/null) || true
    else
        json=$(curl -sL --connect-timeout 10 --max-time 30 "$tags_list_url" 2>/dev/null) || true
    fi

    if [[ -z "$json" ]]; then
        echo "" >&2
        return 1
    fi

    # Parse tags from JSON (prefer jq if available)
    if command -v jq >/dev/null 2>&1; then
        local tags
        tags=$(echo "$json" | jq -r '.tags[]?' 2>/dev/null) || true
        if [[ -n "$tags" ]]; then
            echo "$tags"
            return 0
        fi
    fi
    # Fallback: sed extraction
    local one_line
    one_line=$(echo "$json" | tr -d $'\n')
    if echo "$one_line" | grep -q '"tags"'; then
        echo "$one_line" | sed -E 's/.*"tags"[[:space:]]*:[[:space:]]*\[(.*)\].*/\1/' | tr ',' '\n' | sed 's/^[[:space:]]*"//;s/"[[:space:]]*$//;s/"//g' | tr -d ' '
    else
        echo "" >&2
        return 1
    fi
}

# Filter to version-like tags (semver-ish: digits and dots, optional v prefix).
# Exclude common non-version tags.
filter_version_tags() {
    while read -r tag; do
        [[ -z "$tag" ]] && continue
        # Skip known non-version patterns
        [[ "$tag" == "latest" ]] && continue
        [[ "$tag" == "main" ]] && continue
        [[ "$tag" == "master" ]] && continue
        [[ "$tag" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]] && echo "$tag" && continue
        [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]] && echo "${tag#v}" && continue
        # Allow x.y.z with optional suffix (e.g. 1.0.0-beta)
        [[ "$tag" =~ ^v?[0-9]+\. ]] && echo "$tag" | sed 's/^v//' && continue
    done
}

# Get latest version from OCI registry.
oci_latest_version() {
    local oci_url="$1"
    local tags
    tags=$(oci_list_tags "$oci_url" 2>/dev/null) || return 1
    echo "$tags" | filter_version_tags | sort -V | tail -1
}

# Get latest version from Helm repo (helm search repo repo/chart --versions).
helm_repo_latest_version() {
    local helmrepo="$1"
    local helmrepo_url="$2"
    local repo_name="${helmrepo%%/*}"
    ( helm repo add "$repo_name" "$helmrepo_url" 2>/dev/null; helm repo update "$repo_name" 2>/dev/null ) 1>/dev/null 2>&1 || true
    local versions
    versions=$(helm search repo "$helmrepo" --versions 2>/dev/null | awk 'NR==1 {next} {print $2}')
    echo "$versions" | filter_version_tags | sort -V | tail -1
}

# Compare two version strings (semver-ish). Output: 0 if v1 >= v2, 1 if v1 < v2.
version_lt() {
    local v1="$1"
    local v2="$2"
    local winner
    winner=$(echo -e "${v1}\n${v2}" | sort -V | tail -1)
    [[ "$winner" == "$v2" ]] && [[ "$v1" != "$v2" ]]
}

# Build recommended add-app command for an app.
# For OCI apps: --ocirepo from helmrelease.
# For Helm-repo apps: use mapping (helmrepo, ocipush, helmrepo-url).
recommend_command() {
    local app="$1"
    local new_version="$2"
    local oci_url="$3"
    local use_helmrepo="$4"
    local helmrepo="$5"
    local helmrepo_url="$6"
    local ocipush="$7"

    if [[ "$use_helmrepo" == "true" ]]; then
        if [[ -n "$helmrepo_url" ]]; then
            echo "./catalog-workflow.sh add-app --appname $app --version $new_version --helmrepo $helmrepo --ocipush $ocipush --helmrepo-url $helmrepo_url"
        else
            echo "./catalog-workflow.sh add-app --appname $app --version $new_version --helmrepo $helmrepo --ocipush $ocipush"
        fi
    else
        echo "./catalog-workflow.sh add-app --appname $app --version $new_version --ocirepo $oci_url"
    fi
}

# Process one app: check latest, compare, print recommendation.
check_one_app() {
    local app="$1"
    local latest_catalog
    latest_catalog=$(get_latest_catalog_version "$app")
    if [[ -z "$latest_catalog" ]]; then
        echo -e "  ${YELLOW}$app: no version dirs found${NC}"
        return 0
    fi

    local latest_available=""
    local oci_url=""
    local use_helmrepo=false
    local helmrepo="" helmrepo_url="" ocipush=""

    # Prefer per-app .catalog-source.yaml for Helm-repo apps (pull from Helm repo, push to OCI).
    local source_info
    source_info=$(get_helmrepo_source "$app" 2>/dev/null) || true
    if [[ -n "$source_info" ]]; then
        helmrepo=$(echo "$source_info" | sed -n '1p')
        helmrepo_url=$(echo "$source_info" | sed -n '2p')
        ocipush=$(echo "$source_info" | sed -n '3p')
        if [[ -n "$helmrepo" ]]; then
            use_helmrepo=true
            latest_available=$(helm_repo_latest_version "$helmrepo" "$helmrepo_url") || true
        fi
    fi

    if [[ "$use_helmrepo" != true ]]; then
        oci_url=$(get_oci_url_for_app "$app") || true
        if [[ -z "$oci_url" ]]; then
            echo -e "  ${YELLOW}$app: no OCI URL in helmrelease and no ${CATALOG_SOURCE_FILE} (or job-based/custom)${NC}"
            return 0
        fi
        latest_available=$(oci_latest_version "$oci_url") || true
    fi

    if [[ -z "$latest_available" ]]; then
        echo -e "  ${YELLOW}$app: could not fetch latest version (current in catalog: $latest_catalog)${NC}"
        return 0
    fi

    if version_lt "$latest_catalog" "$latest_available"; then
        echo -e "  ${GREEN}$app: newer version available${NC}"
        echo -e "    Current (catalog): ${CYAN}$latest_catalog${NC}  →  Latest (upstream): ${GREEN}$latest_available${NC}"
        local cmd
        if [[ "$use_helmrepo" == "true" ]]; then
            cmd=$(recommend_command "$app" "$latest_available" "" "true" "$helmrepo" "$helmrepo_url" "$ocipush")
        else
            cmd=$(recommend_command "$app" "$latest_available" "$oci_url" "false")
        fi
        echo -e "    ${BLUE}Run:${NC}"
        echo -e "    ${CYAN}$cmd${NC}"
        echo ""
    else
        echo -e "  ${GREEN}$app: up to date${NC} (catalog: $latest_catalog, upstream: $latest_available)"
    fi
}

# --- Main ---

appname=""
all_apps=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --appname) appname="$2"; shift 2 ;;
        --all) all_apps=true; shift ;;
        -h|--help)
            echo "Usage: $0 [--appname <name>] [--all]"
            echo ""
            echo "Check for latest available chart versions and recommend add-app commands."
            echo ""
            echo "Options:"
            echo "  --appname <name>  Check only this application."
            echo "  --all             Check all applications in applications/."
            echo ""
            echo "Requires: curl. For Helm-repo apps, helm is used to fetch versions."
            exit 0
            ;;
        *) echo -e "${RED}Unknown option: $1${NC}"; exit 1 ;;
    esac
done

if [[ "$all_apps" != true ]] && [[ -z "$appname" ]]; then
    echo -e "${RED}Error: specify --appname <name> or --all${NC}"
    echo "Example: $0 --all"
    echo "Example: $0 --appname podinfo"
    exit 1
fi

echo -e "${BLUE}=== Check latest versions ===${NC}"
echo ""

if [[ "$all_apps" == true ]]; then
    for dir in "${APPLICATIONS_DIR}"/*/; do
        [[ -d "$dir" ]] || continue
        app=$(basename "$dir")
        check_one_app "$app"
    done
else
    if [[ ! -d "${APPLICATIONS_DIR}/${appname}" ]]; then
        echo -e "${RED}Error: application \"${appname}\" not found in applications/${NC}"
        exit 1
    fi
    check_one_app "$appname"
fi

echo -e "${GREEN}Done.${NC}"
