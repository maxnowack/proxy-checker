FROM golang:alpine as build

RUN apk --no-cache add build-base git bzr mercurial gcc curl
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .

RUN go build .
RUN chmod 777 app


FROM alpine
RUN apk --no-cache add ca-certificates
ENV PORT 8080

COPY --from=build /go/src/app/app /

CMD ["/app"]
