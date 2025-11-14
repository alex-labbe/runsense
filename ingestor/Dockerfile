## Multi-stage Dockerfile for the ingestor service
# Builds a static Linux binary and packages it into a small runtime image.

FROM golang:1.25 AS builder
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go env -w GOPATH=/go && go mod download

# Copy the source and build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /out/ingestor ./

FROM alpine:3.18
RUN addgroup -S app && adduser -S -G app app
COPY --from=builder /out/ingestor /usr/local/bin/ingestor
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ingestor"]
