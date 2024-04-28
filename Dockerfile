FROM golang:1.22-alpine as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v ./...

RUN go build -o ./out/go-app .

FROM alpine:3.16.2

COPY --from=build-base /app/out/go-app /go-app
# COPY .env ./.env

EXPOSE 8080

CMD ["./go-app"]
