#!/bin/bash

files=$(git ls-files "*.go" | grep -v vendor | xargs -I {} gofmt -l {})
if [[ -n "${files}" ]]; then
  echo "gofmt needs to be run on the following files:"
  echo "${files}"
  exit 1
fi