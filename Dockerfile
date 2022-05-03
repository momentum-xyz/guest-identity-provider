# syntax=docker/dockerfile:1.3
FROM golang:1.18.1-alpine3.15 as build

WORKDIR /usr/src/code

# Seperate step to allow docker layer caching
COPY go.* ./
RUN go mod download

COPY . ./

RUN go list -m all
RUN go build -o ./build/guest-identity-provider ./cmd/main.go

FROM alpine:3.15.4 as runtime

COPY --from=build /usr/src/code/build/guest-identity-provider /opt/guest-identity-provider

CMD ["/opt/guest-identity-provider"]



