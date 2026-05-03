#!/usr/bin/env bash
# Usage:
#   ./docker-build.sh                        # multi-platform build, no output (CI cache warming)
#   ./docker-build.sh --push                 # multi-platform build and push to registry
#   ./docker-build.sh --local                # single-platform build for current machine, loaded into Docker
#   IMAGE=myregistry.azurecr.io/expirator ./docker-build.sh --push
set -euo pipefail

VERSION="$(tr -d '\r\n' < VERSION)"
IMAGE="${IMAGE:-goodieshq/expirator}"

# Detect local mode
LOCAL=false
PASSTHROUGH_ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "--local" ]]; then
        LOCAL=true
    else
        PASSTHROUGH_ARGS+=("$arg")
    fi
done

# Create a multi-platform builder if one doesn't already exist
if ! docker buildx inspect expirator-builder &>/dev/null; then
    docker buildx create --name expirator-builder --driver docker-container --bootstrap
fi
docker buildx use expirator-builder

if $LOCAL; then
    # Detect host architecture and map to Docker's naming
    case "$(uname -m)" in
        arm64|aarch64) HOST_ARCH="arm64" ;;
        x86_64)        HOST_ARCH="amd64" ;;
        *) echo "Unsupported host arch: $(uname -m)"; exit 1 ;;
    esac

    PLATFORM="linux/${HOST_ARCH}"
    echo "Building ${IMAGE}:local for ${PLATFORM} and loading into Docker..."
    docker buildx build \
        --platform "${PLATFORM}" \
        -t "${IMAGE}:local" \
        --load \
        .
    echo "Done. Run with: docker run --rm ${IMAGE}:local"
else
    PLATFORMS="linux/amd64,linux/arm64"
    echo "Building ${IMAGE}:${VERSION} for ${PLATFORMS}..."
    docker buildx build \
        --platform "${PLATFORMS}" \
        -t "${IMAGE}:${VERSION}" \
        -t "${IMAGE}:latest" \
        "${PASSTHROUGH_ARGS[@]}" \
        .
    echo "Done."
fi
