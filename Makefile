#!/bin/sh

test:
	go test -shuffle on -v -coverprofile coverage.out ./...
	go tool cover -html coverage.out -o coverage.html
