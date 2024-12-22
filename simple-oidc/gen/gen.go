package main

import (
	_ "github.com/ogen-go/ogen"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target api -package api --clean ../../schema.json
