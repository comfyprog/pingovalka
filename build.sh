#!/usr/bin/env bash

VERSION=$1
BIN_PATH=./bin/
mkdir -p $BIN_PATH
for OS in linux windows
do
    for ARCH in 386 amd64 arm arm64
    do
        echo "Building $OS/$ARCH..."
        FILENAME="pingovalka_$OS-$ARCH.$VERSION"
        if [ $OS == "windows" ]; then
            FILENAME="$FILENAME.exe"
        fi
        CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w -X 'main.version=$VERSION'" -o $BIN_PATH/$FILENAME . && upx $BIN_PATH/$FILENAME
    done
done

