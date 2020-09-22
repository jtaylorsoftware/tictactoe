FROM golang:1.14

WORKDIR /go/src/tictactoe
COPY . .

RUN go install -v ./...

FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /usr/local/bin
COPY --from=0 /go/bin/tictactoe .

CMD ["./tictactoe"]