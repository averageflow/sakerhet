#!/bin/sh

coverage-to-html:
	go tool cover -html coverage.out -o coverage.html

execute-test:
	go test -failfast -race -shuffle on -v -coverprofile coverage.out ./...

execute-integration-test:
	export SAKERHET_RUN_INTEGRATION_TESTS=Y; go test -failfast -race -shuffle on -v -coverprofile coverage.out ./...

test: execute-test coverage-to-html

integration-test: execute-integration-test coverage-to-html
