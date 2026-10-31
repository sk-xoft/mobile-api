FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -trimpath -ldflags='-s -w' -o /out/smart-seaman-mobile-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
	&& addgroup -S app \
	&& adduser -S -G app app

ENV TZ=Asia/Bangkok

WORKDIR /app
COPY --from=builder /out/smart-seaman-mobile-api /app/smart-seaman-mobile-api

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/smart-seaman-mobile-api"]
