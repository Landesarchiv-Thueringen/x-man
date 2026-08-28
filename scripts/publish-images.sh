#!/usr/bin/env bash

# Builds and publishes Docker images.
#
# Usage:
#   - Populate your .env file, especially IMAGE_PREFIX and IMAGE_VERSION.
#   - Generally, IMAGE_VERSION should be equal the git tag of the exact commit
#     that is currently checked out. Alternatively, it can be a custom tag like
#     "dev".
#   - Run this script.

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source "$SCRIPT_DIR/../.env"

# temporary bugfix for GitLab
# source: https://forum.gitlab.com/t/started-yesterday-docker-push-error-from-registry-blob-unknown-to-registry/134733
export BUILDX_NO_DEFAULT_ATTESTATIONS=1

tags=(
    gui
    server
    report
)

echo "Build and publish"
for tag in "${tags[@]}"; do
    echo "    $IMAGE_PREFIX/$tag:$IMAGE_VERSION"
done
echo
read -p "Continue? [y/N] "
if [[ $REPLY =~ ^[Yy]$ ]]
then
    echo
    echo Building images...
    (
        cd "$SCRIPT_DIR/.."
        docker compose -f compose.yml build --pull
    )
    echo Publishing images...
    for tag in "${tags[@]}"; do
        docker push "$IMAGE_PREFIX/$tag:$IMAGE_VERSION"
    done
fi
