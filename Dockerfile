
FROM golang:1.22-alpine as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/poll-service
RUN ls -lAtr

FROM alpine:latest
WORKDIR /app

RUN ls -lAtr
COPY --from=builder /build/app app

EXPOSE 8080

CMD ["./app"]