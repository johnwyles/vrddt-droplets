# vrddt

Project skeleton and more importantly architecture from: [spy16/droplets](https://github.com/spy16/droplets)

## Building

vrddt uses `go mod` (available from go 1.11) for dependency management.

To test and build, run `make all`.

## License

TODO
    - Research and implement context correctly
    - CMD
        - ADMIN
        - CLI
        - WORKER
        - API-WEB
            - Authorization / OAuth
            - Frontend JS to use REST API
    - INTERNALS
        - Makefile is broken
        - Dockerize
        - Add S3 storage support
