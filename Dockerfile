FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/multipanel .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=builder /out/multipanel /app/multipanel
COPY config.example.toml /app/config.toml

USER appuser

EXPOSE 8082

ENTRYPOINT ["/app/multipanel", "--config", "/app/config.toml"]
