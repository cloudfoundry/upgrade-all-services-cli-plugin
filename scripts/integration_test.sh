#!/usr/bin/env bash

set -euo pipefail

function main {
  testOutput "Building upgrade-all-services plugin"

  SCRIPTS_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

  BUILD_DIR=`mktemp -d`

  pushd "$SCRIPTS_DIR/.."
      go get -u ./...
      go build -o "$BUILD_DIR/upgrade-all-service-instance-plugin-dev"
  popd

  testOutput "Installing upgrade-all-services plugin"
  cf install-plugin "$BUILD_DIR/upgrade-all-service-instance-plugin-dev" -f

  testOutput "Checking upgrade-all-services plugin is usable"
  cf upgrade-all-services --help

  testOutput "Uninstalling upgrade-all-services plugin"
  cf uninstall-plugin UpgradeAllServiceInstances

  testOutput "Test Success"
}

function testOutput {
  echo -e "\n\n--------\n"$1"\n--------\n\n"
}

main