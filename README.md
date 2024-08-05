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

# Q&A

#### What were some of the key assumptions you made, and what trade-offs did you encounter?

1. I expect a lot of spectators and a hundred persons who vote per second
2. I use single one go routine to send required poll state to the client via websocket. So, it can lead the delays in
   delivery
3. I use redis for vote events distribution as bus, it makes service stateless
4. User management does not exist at all here, so I just imagine users table exists and use person id.
5. During poll creation I don't check same poll existence

#### Imagining this project were to evolve into a full-scale real-world application, what enhancements or next steps would you prioritize to elevate its functionality, user experience, and technical robustness?

* Cover code with tests (unit and integration)
* Add metrics (observability is a must)
* Cache poll ids (to avoid db round trips)
* Cache polls states, keep cache in each service and update them from redis events. (to avoid db round trips)
* Replace redis with kafka/nats/rabbitmq
* Index tables in database
* Rewrite the handler sending the current polls state to the clients. Use go routines to send poll results. New solution
  should properly handle client disconnection and network issues (delays and failures)
* Subscribe websocket clients to the multiple polls at once to reduce connection number
* Use protobuf instead of plain json to reduce amount of traffic server sends
* Implement user management and/or authentication







