# Bytebot Discord Gateway

[![CI](https://github.com/bytebot-chat/gateway-discord/actions/workflows/pull-request.yaml/badge.svg)](https://github.com/bytebot-chat/gateway-discord/actions/workflows/pull-request.yaml)
[![Latest Release](https://img.shields.io/github/v/release/bytebot-chat/gateway-discord?sort=semver)](https://github.com/bytebot-chat/gateway-discord/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/bytebot-chat/gateway-discord)](https://goreportcard.com/report/github.com/bytebot-chat/gateway-discord)
[![Go version](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/bytebot-chat/gateway-discord)
[![Go Reference](https://pkg.go.dev/badge/github.com/bytebot-chat/gateway-discord.svg)](https://pkg.go.dev/github.com/bytebot-chat/gateway-discord)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)
![License](https://img.shields.io/github/license/bytebot-chat/gateway-discord)

## Table of Contents
- [Introduction](#introduction)
- [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Docker](#docker)
    - [Docker-Compose](#docker-compose)
    - [Manual](#manual)
- [Usage](#usage)
    - [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Introduction

_A simpler way to write Discord bots._

Bytebot is a message-passing framework designed to make it easier to write bots for multiple platforms. This repository contains the Discord gateway, which allows you to connect to the Discord API present a single interface to your bot while you run as many applications and services as you like, in any language, without having to worry much about the underlying platform. This particular repository is the gateway for Discord.

This tool is not a complete working bot on its own. The gateway is responsible for authenticating and managing a connection to Discord while passing messages to and from your bot. You will need to write your own bot to take advantage of this gateway. You can see an example of a bot written in Go [here](examples/pingpong/main.go).

Because messages are JSON-encoded, you can write your bot in any language you like. You can even write multiple bots in different languages and run them all at the same time. You can also run multiple instances of the same bot, each with a different configuration, if you like. The only requirement is that your bot can read and write JSON-encoded messages to standard input and output.

Here is an example of a python implementation of a bot that responds to the `ping` command with `pong`: [bytebot-chat/examples/py-pingpong](examples/py-pingpong/pingpong.py)
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

## Contributing

If you would like to contribute to this project, please see the [contributing guidelines](CONTRIBUTING.md).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Acknowledgements

- @m-242 and @parsec for endless code review, suggestions, and tolerating me blowing up their Discord servers with test messages.
- @drewpearce for early iterations and inspiration for the project from a prior project, [Legobot/Legobot](https://github.com/Legobot/).
- Several unnamed friends for their adversarial testing and feedback.