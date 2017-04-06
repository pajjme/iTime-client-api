FROM golang:1.8
RUN mkdir -p /go/src/app

COPY . /go/src/app
WORKDIR /go/src/app

RUN go get -d -v
RUN go build ./pajj_api.go

CMD ["go","run",pajj_api.go"]
