# vrddt

Project skeleton and general architecture from: [spy16/droplets](https://github.com/spy16/droplets)

## TODO

    - Research and implement context correctly
    - CMD
        - ADMIN
        - API
            - Authorization / OAuth
        - CLI
            - GetMetadata
        - WEB
            - Rate limiting
        - WATCHER
        - WORKER
    - INTERNALS
        - Refactor to return errors.XXX instead of fmt.Errorf
        - API Address needs to be sorted out where the Address can be anything local or remote
        - Makefile/Dockerfile/docker-compose.yml refactor for DRY
            - Dockerfile for each command so `server.crt` is only for web and API gcs is only for worker, etc
        - Add S3 storage support
        - Implement other video types for video processor
            - Breakout Upload feature and vrddt video association so it can be
            used again by other types.
