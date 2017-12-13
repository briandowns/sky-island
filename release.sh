#!/bin/sh

NAME="sky-island"
VERSION="0.1"
ARCHS="386 amd64"

echo "Building release..."
for ARCH in ${ARCHS}; do
    BINARY=bin/si-server-${VERSION}-${ARCH}
    GOOS=freebsd GOARCH=${ARCH} go build -v -o ${BINARY}
    tar -czvf bin/${NAME}-${VERSION}-${ARCH}.tgz ${BINARY}
    rm -f ${BINARY}
done

exit 0
