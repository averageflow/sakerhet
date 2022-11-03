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

Unit tests will not run if the special environment variable `SAKERHET_RUN_INTEGRATION_TESTS` is found.

#### Integration tests

Files containing integration tests should be named as `*_integration_test.go`.

Integration tests will be run if the special environment variable `SAKERHET_RUN_INTEGRATION_TESTS` is found.

Integration tests can then be run with `export SAKERHET_RUN_INTEGRATION_TESTS=Y; go test` or equivalent.

This allows a clean separation of test runs. Take a look at the [Makefile](Makefile) in the root of the repo for examples.

There is another special variable, `SAKERHET_INTEGRATION_TEST_TIMEOUT` which configures the timeout of each integration test in seconds. This variable defaults to `60`.
