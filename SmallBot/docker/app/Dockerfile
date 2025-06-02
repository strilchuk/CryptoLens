FROM golang:1.24.2 AS builder

WORKDIR /app

COPY ../../go.mod go.sum ./
RUN go mod download

COPY . .

#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o app .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

#FROM alpine:latest
#
#RUN apk add --no-cache ca-certificates
#
#WORKDIR /root/
#
#COPY --from=builder /app/app .
#COPY --from=builder /app/.env .
#
RUN chmod +x ./app

CMD ["./app"]
