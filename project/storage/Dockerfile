FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o storage .

FROM alpine:3.23 AS final
RUN apk add --no-cache curl
COPY --from=builder /app/storage .
EXPOSE 9999
CMD ["./storage"]