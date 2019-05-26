FROM golang:1.12 AS builder

WORKDIR /appsrc

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o build/app


FROM alpine:latest

WORKDIR /opt

COPY --from=builder /appsrc/build/app .

EXPOSE 8080

CMD ["./app"]
