#!/bin/bash

# Catalog Workflow - Orchestrates add-application, validate, add-tests, build-and-push
# Usage: ./catalog-workflow.sh <command> [options]
#        ./catalog-workflow.sh all [options]
#
# Commands:
#   add-app     Add application (Helm or Kustomize):
#               Helm: --appname, --version, and either --ocirepo or --helmrepo+--ocipush
#               Kustomize: --appname, --version, --kustomize, --gitrepo, --path [--ref]
#   validate    Validate catalog (./validate.sh)
#   add-tests   Create apptest placeholders for app(s)
#   setup       Run setup.sh (ensure tools + go mod tidy for apptests)
#   test        Run apptests (Ginkgo/Kind). Options: [--appname <app>] [--label install|upgrade]
#   build-push     Build and push bundle (needs --tag)
#   all            Run validate + build-push (needs --tag)
#   ci-local       Run CI workflow locally: validate then apptests (same as GitHub Actions)
#   check-versions Check for latest chart versions; recommend add-app commands
#                  Options: --appname <name> | --all
#
# Examples:
#   ./catalog-workflow.sh add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo
#   ./catalog-workflow.sh add-app --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default --ref release-v1.10
#   ./catalog-workflow.sh validate
#   ./catalog-workflow.sh add-tests --appname podinfo
#   ./catalog-workflow.sh add-tests --all
#   ./catalog-workflow.sh test
#   ./catalog-workflow.sh test --appname podinfo
#   ./catalog-workflow.sh setup
#   ./catalog-workflow.sh test --label install
#   ./catalog-workflow.sh build-push --tag v0.1.0
#   ./catalog-workflow.sh all --tag v0.1.0
#   ./catalog-workflow.sh ci-local
#   ./catalog-workflow.sh check-versions --all
#   ./catalog-workflow.sh check-versions --appname podinfo
#   ./catalog-workflow.sh add-app ... --force && ./catalog-workflow.sh validate

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ADD_APP_SCRIPT="${REPO_DIR}/scripts/add-application.sh"
VALIDATE_SCRIPT="${REPO_DIR}/scripts/validate.sh"
BUILD_PUSH_SCRIPT="${REPO_DIR}/scripts/build-and-push.sh"
CHECK_VERSIONS_SCRIPT="${REPO_DIR}/scripts/check-latest-versions.sh"
APPTESTS_DIR="${REPO_DIR}/apptests"

# --- Subcommand implementations ---

cmd_add_app() {
    echo -e "${BLUE}=== add-app ===${NC}"
    [[ -x "$ADD_APP_SCRIPT" ]] || { echo -e "${RED}Missing $ADD_APP_SCRIPT${NC}"; return 1; }
    "$ADD_APP_SCRIPT" "$@"
}

cmd_validate() {
    echo -e "${BLUE}=== validate ===${NC}"
    [[ -x "$VALIDATE_SCRIPT" ]] || { echo -e "${RED}Missing $VALIDATE_SCRIPT${NC}"; return 1; }
    "$VALIDATE_SCRIPT"
}

cmd_build_push() {
    local tag=""
    local args=()
    while [[ $# -gt 0 ]]; do
        case $1 in
            --tag) tag="$2"; shift 2 ;;
            *) args+=("$1"); shift ;;
        esac
    done
    if [[ -z "$tag" ]]; then
        echo -e "${RED}Error: --tag is required for build-push${NC}"
        echo "Example: ./catalog-workflow.sh build-push --tag v0.1.0"
        return 1
    fi
    echo -e "${BLUE}=== build-push (tag=$tag) ===${NC}"
    [[ -x "$BUILD_PUSH_SCRIPT" ]] || { echo -e "${RED}Missing $BUILD_PUSH_SCRIPT${NC}"; return 1; }
    "$BUILD_PUSH_SCRIPT" "$tag"
}

cmd_setup() {
    echo -e "${BLUE}=== setup ===${NC}"
    [[ -x "${REPO_DIR}/scripts/setup.sh" ]] || { echo -e "${RED}Missing scripts/setup.sh${NC}"; return 1; }
    "${REPO_DIR}/scripts/setup.sh"
}

cmd_test() {
    local appname=""
    local label=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --appname) appname="$2"; shift 2 ;;
            --label) label="$2"; shift 2 ;;
            *) shift ;;
        esac
    done

    echo -e "${BLUE}=== test (apptests) ===${NC}"
    if [[ ! -f "${APPTESTS_DIR}/go.mod" ]]; then
        echo -e "${RED}Error: apptests/go.mod not found. Run ./catalog-workflow.sh setup first${NC}"
        return 1
    fi

    if [[ -n "$appname" ]]; then
        echo -e "${GREEN}Running apptests for app: ${appname}${NC} (Ginkgo label filter)"
    fi
    if [[ -n "$label" ]]; then
        echo -e "${GREEN}Label filter: ${label}${NC} (install | upgrade | or custom)"
    fi

    # Build Ginkgo label filter: optional app + optional label (combined with &&)
    local ginkgo_label=""
    if [[ -n "$appname" && -n "$label" ]]; then
        ginkgo_label="${appname} && ${label}"
    elif [[ -n "$appname" ]]; then
        ginkgo_label="$appname"
    elif [[ -n "$label" ]]; then
        ginkgo_label="$label"
    fi

    if command -v just >/dev/null 2>&1; then
        if [[ -n "$ginkgo_label" ]]; then
            just apptests-label "$ginkgo_label"
        else
            just apptests
        fi
    else
        echo -e "${YELLOW}just not found, using go test directly${NC}"
        local args=(-v -timeout 45m)
        [[ -n "$ginkgo_label" ]] && args+=(-ginkgo.label-filter="$ginkgo_label")
        (cd "$APPTESTS_DIR" && go test ./suites/ "${args[@]}")
    fi
}

cmd_add_tests() {
    local appname=""
    local all_apps=false
    while [[ $# -gt 0 ]]; do
        case $1 in
            --appname) appname="$2"; shift 2 ;;
            --all) all_apps=true; shift ;;
            *) shift ;;
        esac
    done

    echo -e "${BLUE}=== add-tests (placeholder scaffolding) ===${NC}"

    if [[ "$all_apps" == true ]]; then
        local apps=()
        for dir in "${REPO_DIR}"/applications/*/; do
            [[ -d "$dir" ]] || continue
            local name
            name=$(basename "$dir")
            apps+=("$name")
        done
        if [[ ${#apps[@]} -eq 0 ]]; then
            echo -e "${YELLOW}No applications found in applications/${NC}"
            return 0
        fi
        echo -e "Creating placeholders for: ${apps[*]}"
        for a in "${apps[@]}"; do
            add_test_placeholder "$a"
        done
    elif [[ -n "$appname" ]]; then
        add_test_placeholder "$appname"
    else
        echo -e "${RED}Error: specify --appname <name> or --all${NC}"
        echo "Example: ./catalog-workflow.sh add-tests --appname podinfo"
        return 1
    fi
    echo -e "${GREEN}✓ Test placeholders created. See docs/APP-TESTS-GUIDE.md for next steps.${NC}"
}

add_test_placeholder() {
    local app="$1"
    local app_dir="${APPTESTS_DIR}/appscenarios"
    local suite_dir="${APPTESTS_DIR}/suites"
    mkdir -p "$app_dir" "$suite_dir"

    local go_name="${app//-/_}"
    local scenario_file="${app_dir}/${go_name}.go"
    local suite_file="${suite_dir}/${go_name}_test.go"

    if [[ -f "$scenario_file" ]]; then
        echo -e "  ${YELLOW}Skip $app: scenario already exists${NC}"
        return 0
    fi

    # PascalCase for struct name (podinfo -> Podinfo, kube-prometheus-stack -> KubePrometheusStack)
    local pascal
    pascal=$(echo "$app" | awk -F- '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2)}1' | tr -d ' ')

    cat > "$scenario_file" << SCENARIO_EOF
//go:build ignore
// +build ignore

// Placeholder: AppScenario for ${app}
// Copy structure from https://github.com/nutanix-cloud-native/nkp-partner-catalog/blob/main/apptests/appscenarios/podinfo.go
// Requires: go.mod, utils/, constant/, github.com/mesosphere/kommander-applications/apptests
// See docs/APP-TESTS-GUIDE.md

package appscenarios

type ${pascal}Placeholder struct{}

func (s *${pascal}Placeholder) Name() string { return "${app}" }
SCENARIO_EOF

    echo -e "  ${GREEN}Created $scenario_file (placeholder - see docs/APP-TESTS-GUIDE.md)${NC}"

    if [[ ! -f "$suite_file" ]]; then
        cat > "$suite_file" << SUITE_EOF
//go:build ignore
// +build ignore

// Placeholder: Ginkgo test suite for ${app}
// Copy from https://github.com/nutanix-cloud-native/nkp-partner-catalog/blob/main/apptests/suites/podinfo_test.go
// See docs/APP-TESTS-GUIDE.md

package suites
SUITE_EOF
        echo -e "  ${GREEN}Created $suite_file (placeholder - see docs/APP-TESTS-GUIDE.md)${NC}"
    fi

    # Create apptests README on first run
    if [[ ! -f "${APPTESTS_DIR}/README.md" ]]; then
        cat > "${APPTESTS_DIR}/README.md" << 'README_EOF'
# App Tests (Placeholders)

Ginkgo/Kind integration tests for catalog applications. Placeholders created by `./catalog-workflow.sh add-tests`.

## Setup

1. Copy `appscenarios/`, `suites/`, `utils/`, `constant/` from [nkp-partner-catalog/apptests](https://github.com/nutanix-cloud-native/nkp-partner-catalog/tree/main/apptests)
2. Or follow [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md) for full setup
3. Replace placeholder `.go` files with real implementations
4. Run: `go test ./suites/ -v -run "<appname>"`
README_EOF
        echo -e "  ${GREEN}Created ${APPTESTS_DIR}/README.md${NC}"
    fi
}

cmd_all() {
    local tag=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --tag) tag="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    if [[ -z "$tag" ]]; then
        echo -e "${RED}Error: --tag required for 'all' (e.g. --tag v0.1.0)${NC}"
        return 1
    fi
    echo -e "${BLUE}=== all (validate + build-push) ===${NC}"
    cmd_validate
    cmd_build_push --tag "$tag"
}

cmd_ci_local() {
    echo -e "${BLUE}=== ci-local (validate + apptests, same as CI) ===${NC}"
    cmd_validate
    echo ""
    cmd_test "$@"
}

cmd_check_versions() {
    echo -e "${BLUE}=== check-versions ===${NC}"
    [[ -x "$CHECK_VERSIONS_SCRIPT" ]] || { echo -e "${RED}Missing $CHECK_VERSIONS_SCRIPT${NC}"; return 1; }
    "$CHECK_VERSIONS_SCRIPT" "$@"
}

show_help() {
    echo "Catalog Workflow - Add app, validate, add tests, build & push"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  add-app      Add application via add-application.sh"
    echo "               Helm: --appname, --version, and either:"
    echo "                 --ocirepo <oci-path>   (chart already in OCI)"
    echo "                 --helmrepo <repo/chart> --ocipush <oci-path> [--helmrepo-url <url>]"
    echo "               Kustomize: --appname, --version, --kustomize --gitrepo <url> --path <path> [--ref <branch|tag>]"
    echo ""
    echo "  validate     Run validate.sh (nkp validate + ghcr.io login)"
    echo ""
    echo "  add-tests    Create apptest placeholders (Ginkgo-style)"
    echo "               Options: --appname <name> | --all"
    echo ""
    echo "  test         Run apptests (Ginkgo/Kind integration tests)"
    echo "               Options: [--appname <app>] [--label install|upgrade]"
    echo ""
    echo "  build-push   Build and push catalog bundle"
    echo "               Options: --tag <version> (e.g. v0.1.0)"
    echo ""
    echo "  all            Run validate + build-push"
    echo "                 Options: --tag <version>"
    echo ""
    echo "  ci-local       Run CI workflow locally (validate then apptests)"
    echo "                 Options: [--appname <app>] [--label install|upgrade] (passed to test)"
    echo ""
    echo "  check-versions Check for latest chart versions; recommend add-app commands"
    echo "                 Options: --appname <name> | --all"
    echo ""
    echo "Examples:"
    echo "  $0 add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo"
    echo "  $0 add-app --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default --ref release-v1.10"
    echo "  $0 validate"
    echo "  $0 add-tests --appname podinfo"
    echo "  $0 add-tests --all"
    echo "  $0 test"
    echo "  $0 test --appname podinfo"
    echo "  $0 test --label install"
    echo "  $0 build-push --tag v0.1.0"
    echo "  $0 all --tag v0.1.0"
    echo ""
    echo "Run multiple: $0 add-app ... && $0 validate && $0 build-push --tag v0.1.0"
}

# --- Main ---

COMMAND="${1:-}"
shift || true

case "$COMMAND" in
    add-app)
        cmd_add_app "$@"
        ;;
    validate)
        cmd_validate "$@"
        ;;
    add-tests)
        cmd_add_tests "$@"
        ;;
    setup)
        cmd_setup "$@"
        ;;
    test)
        cmd_test "$@"
        ;;
    build-push)
        cmd_build_push "$@"
        ;;
    all)
        cmd_all "$@"
        ;;
    ci-local)
        cmd_ci_local "$@"
        ;;
    check-versions)
        cmd_check_versions "$@"
        ;;
    -h|--help|help|"")
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        show_help
        exit 1
        ;;
esac
