#!/bin/bash

# exit when a command fails
set -e

echo "Beginning containerz build"
blaze build --config=nocgo :containerz
echo "containerz build complete, beginning docker build"

CONTAINERZ_BINARY="$(cat "$(blaze info master-log)" | grep output_file | awk '{print $3}')"

# Create temporary directory and copy in binary.
mkdir build-out
cp "$CONTAINERZ_BINARY" ./build-out/

docker build -t containerz -f Dockerfile .

# Remove the temporary build directory.
rm -rf ./build-out/

echo "docker build complete. Have a nice day."
