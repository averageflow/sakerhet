#!/bin/sh

coverage-to-html:
	go tool cover -html coverage.out -o coverage.html

execute-test:
	go test -race -shuffle on -v -coverprofile coverage.out ./...

execute-integration-test:
	go test -race -shuffle on -v -coverprofile coverage.out --tags=integration ./...

test: execute-test coverage-to-html

integration-test: execute-integration-test coverage-to-html
