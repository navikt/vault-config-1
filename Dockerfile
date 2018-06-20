# build stage
FROM golang:1.10.1-alpine3.7 AS build-env

WORKDIR /go/src/github.com/elliottsam/vault-config
COPY version version/
COPY vault vault/
COPY template template/
COPY main.go ./
COPY crypto crypto/
COPY cmd cmd/
COPY vendor vendor/
RUN ls -lt  /go/src/github.com/elliottsam/vault-config
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vault-config  

# Image to build 
FROM alpine:3.7

WORKDIR /app
COPY --from=build-env go/src/github.com/elliottsam/vault-config/vault-config . 
RUN chmod 755 vault-config
CMD ["--help"]
ENTRYPOINT ["./vault-config"]
