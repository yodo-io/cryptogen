# build stage
FROM golang:1.11-alpine3.8 AS builder
ENV PORT 8080
EXPOSE 8080

RUN apk --no-cache add git && go get -u github.com/golang/dep/cmd/dep

# install deps first to make better use of layer cache
WORKDIR /go/src/github.com/yodo-io/cryptogen
COPY Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only=true

COPY . .
RUN go build -o /bin/cryptogen

CMD ["/bin/cryptogen"]

# run stage
FROM alpine:3.8  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/cryptogen /bin/
CMD ["/bin/cryptogen"]
