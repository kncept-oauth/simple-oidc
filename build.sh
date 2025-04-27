#!/bin/bash
set -euo pipefail

clean() {
    echo do clean
}

prepare() {
  cd service 
  go generate gen/gen.go
  cd ..
}

build() {
    echo do build

    cd service 
    go generate gen/gen.go  
    cd ..
}

mod() {
    echo go mod
}

testharness() {
    cd testharness && go run main.go
}

if [[ $# -eq 0 ]] ; then
  echo targets: clean prepare build testharness
  exit 1
fi

while [[ $# > 0 ]] ; do
  case "$1" in
    clean)
      (clean)
      ;;
    prepare)
      (prepare)
      ;;
    build)
      (build)
      ;;
    testharness)
      (testharness)
      ;;
    *)
      echo Unknown option: $1
      exit 1
      ;;
  esac
  shift
done