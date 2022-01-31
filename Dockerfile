FROM golang:1.17

WORKDIR /src
COPY . /src

RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o output/gateway github.com/moycat/shiba-nat/cmd/gateway
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o output/client github.com/moycat/shiba-nat/cmd/client

FROM debian:11-slim

COPY entrypoint.sh /entrypoint.sh
COPY --from=0 /src/output/gateway /usr/bin/shiba-nat-gateway
COPY --from=0 /src/output/client /usr/bin/shiba-nat-client
ENTRYPOINT ["/entrypoint.sh"]
