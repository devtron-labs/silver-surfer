# Binary Build
FROM golang:1.21-alpine3.17  AS build-env
RUN echo $GOPATH
RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN mkdir /silver-surfer
WORKDIR /silver-surfer
ADD . /silver-surfer/
#ARG AUTH_TOKEN
#RUN test -n "$AUTH_TOKEN"
#ENV GITHUB_TOKEN=${AUTH_TOKEN}
#ARG RELEASE
#RUN if [ "$RELEASE" = "goreleaser" ]; then echo `make release`; fi
RUN GOOS=linux make

# Prod Build
FROM alpine3.17
RUN apk add --no-cache ca-certificates
RUN apk update
RUN apk add git
# RUN if [ "$RELEASE" = "goreleaser" ]; then echo `make release`; fi
COPY --from=build-env  /silver-surfer/bin .
ENTRYPOINT ["./kubedd"]
