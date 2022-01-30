FROM golang:1.17

ADD ./src/ /src
WORKDIR /src

RUN mkdir bin
ENV PATH ${GOPATH}/bin:${PATH}

RUN go get github.com/fulldump/box
RUN go get github.com/globalsign/mgo
RUN go build -o server *.go

EXPOSE 8080

ARG BENCHMARK_ENV
ARG TFB_TEST_DATABASE
ARG TFB_TEST_NAME

CMD ./server
