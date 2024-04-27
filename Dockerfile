FROM golang:1.20.7 as builder

WORKDIR /opt/app
ARG GH_USER
ARG GH_TOKEN
RUN git config --global url.https://${GH_USER}:${GH_TOKEN}@github.com/.insteadOf https://github.com/

COPY ./tools ./tools/
RUN make -f ./tools/Makefile tools  # tools cache layer

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM debian:bullseye-slim

RUN apt update && \
    apt install -y ca-certificates && \
    apt clean

COPY --from=builder /opt/app/bin/* /usr/local/bin/
