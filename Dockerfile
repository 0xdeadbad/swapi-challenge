FROM golang:1.21.4-alpine3.18 AS builder

ADD . /swapi
WORKDIR /swapi
RUN CGO_ENABLED=0 GOOS=linux go build -o /swapi/swapi

FROM alpine:3.18

COPY --from=builder /swapi/swapi /bin/swapi

EXPOSE 8080

ENTRYPOINT [ "/bin/swapi" ]