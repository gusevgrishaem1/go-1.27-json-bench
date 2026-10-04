#!/bin/sh
set -eu

benchtime=${1:-2s}
count=${2:-5}
case "$count" in
    ''|*[!0-9]*|0) echo 'Count must be a positive integer' >&2; exit 1 ;;
esac

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
results="$project_dir/benchmark-results"
mkdir -p "$results"
export GOCACHE="$project_dir/.go-build-cache"

{
    go version
    uname -a
    printf 'benchtime=%s count=%s\n' "$benchtime" "$count"
    printf '%s\n' 'Small=1 item; Large=1000 items; Encode=Response; Decode=fresh Request'
} > "$results/environment.txt"

# Capture first, then print: a failing go test must not be hidden by tee.
run_variant() {
    variant=$1
    module=$2
    experiment=$3
    printf '%s: GOEXPERIMENT=%s\n' "$variant" "$experiment"
    if (cd "$project_dir/$module" && GOEXPERIMENT="$experiment" go test ./... \
        -run '^$' -bench '^BenchmarkJSON$' -benchmem \
        "-benchtime=$benchtime" "-count=$count") > "$results/$variant.txt" 2>&1; then
        cat "$results/$variant.txt"
    else
        cat "$results/$variant.txt" >&2
        return 1
    fi
}

run_variant go127-v1 go127-v1 jsonv2
run_variant go127-v1-nojsonv2 go127-v1 nojsonv2
run_variant go127-v2 go127-v2 jsonv2
