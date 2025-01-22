#!/bin/bash

ALL_CHANGED_FILES="iso/country/country.go iso/country/go.mod iso/go.mod iso/site/README iso/site/go.mod iso/site/sites.go .github/workflows/check.yml currency/currency.go currency/currency_test.go currency/go.mod currency/go.sum"
# get all the changed directories
changed_dirs=$(echo "$ALL_CHANGED_FILES" | tr ' ' '\n' | xargs -I {} dirname {} | sort -u | uniq)
# get all changed packages
changed_pkgs=()
for dir in $changed_dirs; do
    # check if the dir has go.mod in it, recursively
    updated_dirs=$(find "$dir" -name 'go.mod' -exec dirname {} \; | sort -u | uniq)
    for updated_dir in $updated_dirs; do
        changed_pkgs+=("$updated_dir")
    done
done
# unique the changed packages
IFS=" " read -r -a changed_pkgs <<<"$(tr ' ' '\n' <<<"${changed_pkgs[@]}" | sort -u | tr '\n' ' ')"

revive_config=$(realpath revive.toml)
for pkg in "${changed_pkgs[@]}"; do
    # check do we need to run revive, include go files
    have_go_files=$(find "$pkg" -depth 1 -name '*.go' -exec dirname {} \; | sort -u | uniq)
    if [ -z "$have_go_files" ]; then
        continue
    fi
    echo "running go vet and revive for $pkg"
    pushd "$pkg" >/dev/null 2>&1 || exit 255
    go vet ./... || exit 255
    revive -formatter friendly -config $revive_config ./... || exit 255
    popd >/dev/null 2>&1 || exit 255
done
