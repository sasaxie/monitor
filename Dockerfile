FROM golang:1.11 as builder
WORKDIR /go/src/github.com/sasaxie/monitor
COPY . /go/src/github.com/sasaxie/monitor
RUN go install .

CMD ["monitor"]