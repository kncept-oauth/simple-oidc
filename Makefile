MAKEFLAGS += --silent

generate:
	cd service && go generate gen/gen.go

build: generate
	cd service &&  go build -o simple-oidc main.go 

cdkbuild: generate
	mkdir -p deploy/bin
	cd service && go build -o ../deploy/bin/bootstrap main.go 

.PHONY: testharness
testharness:
	cd testharness && go run main.go

.PHONY: deployable
deployable: cdkbuild
	cd deploy && npm run cdk ls


deploy: cdkbuild
	cd deploy && npm run cdk ls
	echo "cd deploy && npm run cdk deploy simple-oidc"

help:
	echo TODO: print help