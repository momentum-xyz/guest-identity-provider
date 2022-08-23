# syntax=docker/dockerfile:1.3
FROM golang:1.19.0-alpine3.16 as build

WORKDIR /usr/src/code

# Seperate step to allow docker layer caching
COPY go.* ./
RUN go mod download

COPY . ./

RUN go list -m all
RUN go build -o ./build/guest-identity-provider ./cmd/main.go

FROM alpine:3.16.1 as runtime

COPY --from=build /usr/src/code/build/guest-identity-provider /opt/guest-identity-provider

CMD ["/opt/guest-identity-provider"]



