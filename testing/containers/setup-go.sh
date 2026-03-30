#!/bin/bash
set -e

gotar="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
gourl="https://dl.google.com/go/${gotar}"

echo "Downloading Go ${GO_VERSION} for ${GOARCH}..."
curl -o "/tmp/${gotar}" "${gourl}"
tar -x -C /opt/ -f "/tmp/${gotar}"
rm -f "/tmp/${gotar}"
echo "Go ${GO_VERSION} installed successfully"
