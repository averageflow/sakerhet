# Säkerhet

_Säkerhet_ - from Swedish, meaning security, certainty

Helpful abstractions to ease the creation of integration tests in Go with some key ideas:

- Tests should be easy to **read, write, run and maintain**
- Using Docker containers with real instances of services (PostgreSQL, NGINX, GCP Pub/Sub, etc) with testcontainers
- Integrate with the standard Go library
- Encourage loose coupling between parts

Tests are a great source of documentation and great way of getting to know a project.

As the business requirement or documentation changes, the test case will also adjust according to the needs. This will make the test case good documentation for the developer and will increase the confidence of the developer when refactoring or doing something.

![safety-net](https://user-images.githubusercontent.com/42377845/198402912-d9cf2925-6a7b-4f5c-9709-e1f24a9f827b.jpg)

## Docs

### Test separation

This is not a strict requirement but a good practice. We encourage you to keep the 2 types of test separated.

There are plenty of reasons developers might want to separate tests. Some simple examples might be:

- Integration tests are often slower, so you may want to only run them after the unit test (which are often much faster) have passed.
- Smoke tests that are run against the live application, generally after a deployment.
- Deploying the same app to different tenants.

Of course, this is by no means an exhaustive list!

#### Unit tests

Files containing unit test should be named as `*_test.go`.

Unit test files should contain in the first lines of the file the following snippet, to avoid being run on integration test runs:

```go
//go:build !integration
```

#### Integration tests

Files containing integration tests should be named as `*_integration_test.go`.

Integration test files should contain in the first lines of the file the following snippet, to avoid being run on normal test runs:

```go
//go:build integration
```

Integration tests can then be run with `go test --tags=integration` or equivalent.

This allows a clean separation of test runs. Take a look at this example Makefile, or at the one in the root of the repo for examples:

```Makefile
#!/bin/sh

coverage-to-html:
 go tool cover -html coverage.out -o coverage.html

execute-test:
 go test -race -shuffle on -v -coverprofile coverage.out ./...

execute-integration-test:
 go test -race -shuffle on -v -coverprofile coverage.out --tags=integration ./...

test: execute-test coverage-to-html

integration-test: execute-integration-test coverage-to-html
```
