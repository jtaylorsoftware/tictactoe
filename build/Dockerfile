FROM golang:1.14

WORKDIR /go/src/tictactoe
COPY . .

RUN go install -v .

CMD ["tictactoe"]