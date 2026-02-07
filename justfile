# Justfile for NKP catalog — validate, apptests, build
# Install just: https://just.systems/man/en/chapter_3.html
# Usage: just [recipe] [arguments]

_repo_root := justfile_directory()
_apptests_dir := _repo_root + "/apptests"

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
    go test ./suites/ -v

# Run apptests for a specific app
# Usage: just apptests-app podinfo
apptests-app app:
    #!/usr/bin/env bash
    set -e
    cd "{{ _apptests_dir }}"
    if [ ! -f go.mod ]; then
        echo "apptests/go.mod not found. Run setup per docs/APP-TESTS-GUIDE.md"
        exit 1
    fi
    go test ./suites/ -v -run "{{ app }}"

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
    go test ./suites/ -v -ginkgo.label-filter="upgrade"

# Ensure apptests dependencies are tidy
apptests-tidy:
    cd "{{ _apptests_dir }}" && go mod tidy

# Alias: test = apptests
test: apptests
