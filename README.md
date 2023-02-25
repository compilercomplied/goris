# Goris

Redis implementation in golang for educational purposes.

# Running

There are executables available under the `cmd` dir. One for a test client and another one for the server.
```sh
# On the root dir.

go run ./cmd/client

go run ./cmd/server
```

# Development

To execute the tests run `go test -v ./...` to recursively expand on subdirs. Verbosity flag provides information about the tests that have been run.
