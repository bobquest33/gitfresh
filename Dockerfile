FROM golang:1.4

RUN mkdir -p /go/src/github.com/syb-devs/gitfresh

COPY . /go/src/github.com/syb-devs/gitfresh

WORKDIR /go/src/github.com/syb-devs/gitfresh

RUN go get 
RUN go build .

ENTRYPOINT ["gitfresh", "-path", "/var/git/data"]
CMD ["-h"]
