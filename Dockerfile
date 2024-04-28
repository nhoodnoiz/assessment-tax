# FROM golang:1.22-alpine as build-base

# WORKDIR /app

# COPY go.mod .

# RUN go mod download

# COPY . .

# RUN CGO_ENABLED=0 go test -v ./...

# RUN go build -o ./out/go-app .

# FROM alpine:3.16.2

# COPY --from=build-base /app/out/go-app /app/go-app

# COPY .env /app/.env

# EXPOSE 8080

# WORKDIR /app

# ENV DATABASE_URL="host=localhost port=5432 user=myuser password=mypassword dbname=mydatabase sslmode=disable"

# ENV PORT=8080

# ENV ADMIN_USERNAME=adminTax
# ENV ADMIN_PASSWORD=admin!

# CMD ["/app/go-app"]

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


ENV DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=ktaxes sslmode=disable"
ENV PORT=8080
ENV ADMIN_USERNAME=adminTax
ENV ADMIN_PASSWORD=admin!


CMD ["./go-app"]
