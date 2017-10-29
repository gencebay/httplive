FROM golang:1.9.2 as builder
WORKDIR /go/src/github.com/gencebay/httpbin/
RUN go get -d -v github.com/gin-gonic/gin  
COPY .    .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/gencebay/httpbin/app .
COPY --from=builder /go/src/github.com/gencebay/httpbin/wwwroot ./wwwroot
CMD ["./app"]  