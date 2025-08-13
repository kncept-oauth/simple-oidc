#!/bin/bash
set -euo pipefail

clean() {
    cd service 
    go clean
    cd ..

    cd testharness
    go clean
    cd ..
}

prepare() {
  cd service 
  go mod tidy
  go generate gen/gen.go
  cd ..
  
  cd testharness
  go mod tidy
  cd ..

  cd deploy
  npm ci
  cd ..
}

test() {
  cd service 
  go test ./...
  go test -tags integration ./...
  cd ..
  
  cd testharness 
  go test ./...
  cd ..
  

  # cd deploy
  # npm run test
  # cd ..
}

build() {
    cd service 
    go generate gen/gen.go  
    CGO_ENABLED=0 GOOS=linux GOARGS=amd64 go build -o bootstrap -ldflags="-s -w"
    cd ..
    
}

testharness() {
    cd testharness && go run main.go
}

deploy() {
  cd deploy
  npm run cdk deploy
  cd ..
}

if [[ $# -eq 0 ]] ; then
  echo targets: clean prepare test build testharness deploy
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
    test)
      (test)
      ;;
    testharness)
      (testharness)
      ;;
    deploy)
      (deploy)
      ;;
    *)
      echo Unknown option: $1
      exit 1
      ;;
  esac
  shift
done
