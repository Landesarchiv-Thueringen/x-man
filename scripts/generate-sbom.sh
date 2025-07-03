#!/usr/bin/env bash

# Go to repository root dir
cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.."

tags=(
    gui
    server
    report
)

docker compose -f compose.yml build --pull
for tag in "${tags[@]}"; do
    docker scout sbom --format list "localhost/x-man/$tag" > "docs/sbom/sbom_x-man_$tag.txt"
    docker scout sbom --format cyclonedx "localhost/x-man/$tag" > "docs/sbom/sbom_x-man_$tag.cdx"
done