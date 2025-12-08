FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /app/bin/gateway ./cmd/app

FROM alpine:3.23 AS final
RUN apk add --no-cache curl
COPY --from=builder /app/bin/gateway /usr/local/bin/gateway
EXPOSE 8080 9090
CMD ["/usr/local/bin/gateway"]