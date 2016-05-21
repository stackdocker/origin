#!/bin/bash -e

ORIGIN_ROOT=$(dirname "${BASH_SOURCE}")

if [[ ! -d "$ORIGIN_ROOT/_output/local/bin/linux/amd64/" ]]; then
  echo "please make first"
  exit 1
fi

cp -u $ORIGIN_ROOT/Dockerfile $ORIGIN_ROOT/_output/local/bin/linux/amd64/

docker build -t tangfeixiong/openshift-origin $ORIGIN_ROOT/_output/local/bin/linux/amd64/
