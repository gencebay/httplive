FROM golang:1.15.5-alpine as builder

WORKDIR /src/app

ENV CGO_ENABLED=0

COPY . .

RUN GOOS=linux go build -a -installsuffix cgo -o httplive .

FROM alpine:3.7

WORKDIR /src/app

EXPOSE 5003

VOLUME /src/app

ENV GIN_MODE release

COPY --from=builder /src/app/httplive .

RUN chmod +x /src/app/httplive

ENTRYPOINT ["./httplive"]