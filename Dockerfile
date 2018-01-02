FROM golang:1.9.2 as builder
WORKDIR /go/src/github.com/gencebay/httplive/
RUN go get -d -v github.com/gin-gonic/gin
RUN go get -d -v github.com/boltdb/bolt
RUN go get -d -v github.com/gin-gonic/contrib/static
RUN go get -d -v github.com/urfave/cli
COPY .    .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM nginx:alpine
ENV APPDIRPATH /go/src/github.com/gencebay/httplive/
RUN apk --no-cache add ca-certificates
WORKDIR ${APPDIRPATH}
COPY --from=builder ${APPDIRPATH}/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder ${APPDIRPATH}/nginx.vh.default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder ${APPDIRPATH}/app .
COPY --from=builder ${APPDIRPATH}/public ./public
COPY --from=builder ${APPDIRPATH}/dockerstart.sh ./dockerstart.sh

# Expose ports
EXPOSE 80

#CMD ["./app"]
CMD ["sh","dockerstart.sh"]