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
