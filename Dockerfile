FROM golang:1.18.8-alpine3.16 AS builder

RUN adduser -D -g 'bytebot' bytebot
WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags "-s -w -extldflags '-static'" -o ./opt/bytebot

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/opt/bytebot /opt/bytebot
VOLUME /data

# Our chosen default for Prometheus
EXPOSE 8080
USER bytebot
ENTRYPOINT ["/opt/bytebot"]
