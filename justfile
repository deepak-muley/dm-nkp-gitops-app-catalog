# Justfile for NKP catalog — validate, apptests, build
# Install just: https://just.systems/man/en/chapter_3.html
# Usage: just [recipe] [arguments]

_repo_root := justfile_directory()
_apptests_dir := _repo_root + "/apptests"
_catalog_apptests_dir := _repo_root + "/catalog-apptests"
# Apptests run Kind + Flux + app deploy; cluster bring-up (e.g. MetalLB) can be slow
_apptests_timeout := "45m"

# Run same steps as CI locally: validate then apptests (see docs/TEST-CI-LOCALLY.md)
ci-local:
    cd "{{ _repo_root }}" && ./catalog-workflow.sh ci-local

# Run all apptests (Ginkgo/Kind integration tests)
# Requires: apptests/ with go.mod, full setup per docs/APP-TESTS-GUIDE.md
apptests:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    go test ./suites/ -v -timeout "{{ _apptests_timeout }}"

# Run apptests for a specific app (filters by Ginkgo label, e.g. Label("podinfo"))
# Usage: just apptests-app podinfo
apptests-app app:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    echo "Running apptests for app: {{ app }}"
    go test ./suites/ -v -timeout "{{ _apptests_timeout }}" -ginkgo.label-filter="{{ app }}"

# Run apptests with a Ginkgo label filter (e.g. "install", "upgrade", "podinfo && install")
# Usage: just apptests-label "install"   |  just apptests-label "podinfo && upgrade"
apptests-label label_filter:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    echo "Running apptests with label filter: {{ label_filter }}"
    go test ./suites/ -v -timeout "{{ _apptests_timeout }}" -ginkgo.label-filter='{{ label_filter }}'

# Run apptests with Ginkgo label filter (install tests only)
apptests-install:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    go test ./suites/ -v -ginkgo.label-filter="install"

# Run apptests with Ginkgo label filter (upgrade tests only)
apptests-upgrade:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    go test ./suites/ -v -timeout "{{ _apptests_timeout }}" -ginkgo.label-filter="upgrade"

# Run catalog-apptests (all apps under applications/ — no per-app test code)
# Usage: just apptests-templated
apptests-templated:
    #!/usr/bin/env bash
    set -e
    cd "{{ _catalog_apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "catalog-apptests/go.mod not found."
        exit 1
    fi
    go test . -v -timeout "{{ _apptests_timeout }}"

# Run catalog-apptests for one app (label appname=<app>)
# Usage: just apptests-templated-app podinfo
apptests-templated-app app:
    #!/usr/bin/env bash
    set -e
    cd "{{ _catalog_apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "catalog-apptests/go.mod not found."
        exit 1
    fi
    echo "Running catalog-apptests for app: {{ app }}"
    go test . -v -timeout "{{ _apptests_timeout }}" -ginkgo.label-filter="appname={{ app }}"

# Run catalog-apptests with a label filter (e.g. "install", "appname=podinfo && upgrade")
# Usage: just apptests-templated-label "install"
apptests-templated-label label_filter:
    #!/usr/bin/env bash
    set -e
    cd "{{ _catalog_apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "catalog-apptests/go.mod not found."
        exit 1
    fi
    echo "Running catalog-apptests with label filter: {{ label_filter }}"
    go test . -v -timeout "{{ _apptests_timeout }}" -ginkgo.label-filter='{{ label_filter }}'

# Ensure apptests dependencies are tidy
apptests-tidy:
    cd "{{ _apptests_dir }}" && go mod tidy

# Alias: test = apptests
test: apptests
