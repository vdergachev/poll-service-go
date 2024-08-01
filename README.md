# poll-service-go
A RESTful service in Go that enables users to create polls, vote in real-time, and view poll results as they come in.

# Build

## Prerequisites

You have to install docker, go 1.22 and make tool before build the application

* docker
* golang
* make

## MacOS

```shell
make build
```

## Linux

```shell
make build_linux
```

# Usage

## Dependencies

Start database and redis before run the app

```shell
docker-compose up -d
```

## Run application

Go to the project root and execute following command

```shell
./poll-service
```

## Create Poll

```shell
curl -X POST -H "Content-Type: application/json" -d @poll.json http://localhost:4000/polls
```

## Subscribe

```shell
curl -i -N -H "Upgrade: websocket" -H "Connection: Upgrade" -H "Sec-WebSocket-Key: $(openssl rand -base64 16)" -H "Sec-WebSocket-Version: 13" http://localhost:4000/ws/polls/1
```

## Vote

```shell
curl -X PUT http://localhost:4000/polls/1/options/2/users/3
```
