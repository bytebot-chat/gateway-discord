# Bytebot Discord Gateway

[![Docker Build and Push](https://github.com/bytebot-chat/gateway-discord/actions/workflows/docker_build.yaml/badge.svg?branch=0.0.1)](https://github.com/bytebot-chat/gateway-discord/actions/workflows/docker_build.yaml)
![Latest Release](https://img.shields.io/github/v/release/bytebot-chat/gateway-discord?sort=semver&style=plastic)

## Introduction

This is the Discord gateway for the Bytebot chat ecosystem. It is meant to be a relatively simple gateway that can be used to connect to Discord and relay messages to and from the Bytebot ecosystem. By using this gateway, you can connect all of your apps to Discord without having to write a Discord bot for each one or managing multiple sets of credentials. Business logic is handled by the Bytebot ecosystem, not the gateway, so you can focus on building your app and not on the Discord API.

## Installation

### Prerequisites
- Redis server (for pub/sub)
- Docker (optional, required for docker or docker-compose)
- Golang (optional, required for building from source)
- A Discord bot token (see [here](https://discord.com/developers/docs/intro) for more info)

### Docker

The easiest way to get started is to use the Docker image. You can find the image on the Github Container Registry [here](https://github.com/bytebot-chat/gateway-discord/pkgs/container/gateway-discord). It is strongly recommended that you pull either the `main` or `edge` tag, as these are the most up-to-date versions of the gateway. Semver tags are only updated when a new release is made but are not guaranteed to be the most up-to-date version. However, they are guaranteed to be more stable.

### Docker Compose

The easiest way to get started for development or just poking around is to use the provided [example docker-compose file](docker-compose-example.yaml). This file will start the gateway and a Redis server used for pub/sub and is required for the gateway to function. You will also need to provide your Discord bot token in the args section of the gateway service in the `docker-compose-example.yml` file. You can see your existing apps and retrieve your Discord bot token here [here](https://discord.com/developers/applications).

1. Copy the `docker-compose-example.yml` file to `docker-compose.yml`
2. Edit the `docker-compose.yml` file and replace the `YOUR_DISCORD_BOT_TOKEN` placeholder with your Discord bot token
3. Run `docker-compose up -d` to start the gateway and Redis server.

From this point you can connect to the redis pubsub to watch for messages from the gateway or connect an app. Try running one of the apps in [`examples/`](examples/) to see how it works.

### Manual

If you would like to install the gateway manually, you can do so by cloning this repository and running `go build` in the root directory. You will need to have Go installed on your system. You can find instructions for installing Go [here](https://golang.org/doc/install). 

You may also install the gateway using `go get`:

```bash
go get github.com/bytebot-chat/gateway-discord
```

To install a specific version, you can use the `@` symbol to specify a version:

```bash
go get github.com/bytebot-chat/gateway-discord@v0.1.0
```

## Usage

### Configuration

The gateway can currently only be configured by CLI flags. The following table shows the available configuration options:

| Flag | Description | Default |
| --- | --- | --- |
| redis | The address of the Redis server | `localhost:6379` |
| rpass | The password for the Redis server | `""` |
| token | The Discord bot token | `""` |
| id | The ID of the Discord bot | `""` |
| inbound | The channel to listen for inbound messages on | `discord-inbound` |
| outbound | The channel to publish outbound messages to | `discord-outbound` |
| verbose | Whether to print verbose logs | `false` |

### Contributing

If you would like to contribute to this project, please see the [contributing guidelines](CONTRIBUTING.md).

### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

### Acknowledgements

- @m-242 and @parsec for endless code review, suggestions, and tolerating me blowing up their Discord servers with test messages.
- @drewpearce for early iterations and inspiration for the project from a prior project, [Legobot/Legobot](https://github.com/Legobot/).
- Several unnamed friends for their adversarial testing and feedback.