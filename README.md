# vrddt

Project skeleton and general architecture from: [spy16/droplets](https://github.com/spy16/droplets)

## TODO

    - Research and implement context correctly
    - CMD
        - ADMIN
        - API
            - Authorization / OAuth
            - Rate limiting
        - CLI
            - Get Metadata
        - WEB
            - Authorization / OAuth
            - Rate limiting
        - WATCHER
        - WORKER
    - INTERNALS
        - API Address needs to be sorted out where the Address can be anything local or remote
        - Makefile/Dockerfile/docker-compose.yml refactor for DRY
        - Add S3 storage support
        - Implement other video types for video processor
            - Breakout Upload feature and vrddt video association so it can be
            used again by other types.
