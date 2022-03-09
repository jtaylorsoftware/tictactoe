FROM golang:1.17 AS base

WORKDIR /usr/src/tictactoe

# download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# copy over sources
COPY . .

FROM base AS test

ENTRYPOINT ["go", "test", "./..."]

FROM base AS build

RUN go install -v ./...

FROM alpine:latest AS prod
RUN apk add --no-cache libc6-compat
WORKDIR /usr/local/bin
COPY --from=build /go/bin/tictactoe .

ENTRYPOINT ["./tictactoe"]