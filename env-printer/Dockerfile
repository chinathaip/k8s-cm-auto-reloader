FROM golang:1.19-alpine as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v

RUN go build -o ./out/kbtg-go-assesment .


# ====================
FROM alpine:3.16.2
COPY --from=build-base /app/out/kbtg-go-assesment /app/kbtg-go-assesment

CMD ["/app/kbtg-go-assesment"]