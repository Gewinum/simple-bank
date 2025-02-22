# Use the official Golang image as the base image
FROM golang:1.23.4-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

RUN apk --no-cache add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o app ./cmd/app/main.go

# Start a new stage from scratch
FROM alpine:3.21

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add curl

COPY --from=builder /app/app /app
COPY --from=builder /app/migrate /migrate

COPY app.env .
COPY db/migration ./migration
COPY scripts/start.sh ./start.sh
# Expose port 80 to the outside world
EXPOSE 80

# Command to run the executable
CMD ["/app"]
ENTRYPOINT ["./start.sh"]