FROM golang:1.8
RUN mkdir -p /go/src/github.com/pajjme/iTime-client-api/

COPY . /go/src/github.com/pajjme/iTime-client-api/
WORKDIR /go/src/github.com/pajjme/iTime-client-api/

RUN sh deps.sh
RUN go build -o app ./pajj_api.go
CMD ["./app"]
