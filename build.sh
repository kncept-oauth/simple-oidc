#!/bin/bash
set -euo pipefail

cat << EOF
Currently moving from make to bash
EOF

clean() {
    echo do clean
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


while [[ $# > 0 ]] ; do
  case "$1" in
    clean)
      clean
      ;;
    prepare)
      prepare
      ;;
    build)
      build
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