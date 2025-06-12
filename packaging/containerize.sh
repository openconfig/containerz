#!/bin/bash

# exit when a command fails
set -e

docker build -t containerz ../

echo "docker build complete. Have a nice day."
