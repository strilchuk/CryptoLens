FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /app

COPY ../../go.mod go.sum ./
RUN go mod download

COPY . .

# Определяем архитектуру для сборки
RUN case ${TARGETPLATFORM} in \
    "linux/amd64") \
        GOARCH=amd64 ;; \
    "linux/arm/v7") \
        GOARCH=arm GOARM=7 ;; \
    "linux/arm64") \
        GOARCH=arm64 ;; \
    *) \
        echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} GOARM=${GOARM:-} go build -o app .

FROM --platform=$TARGETPLATFORM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/.env .

RUN chmod +x ./app

CMD ["./app"]
