FROM golang as build


COPY . /go/src/github.com/zhangkesheng/registrator

ENV CGO_ENABLED=0

RUN cd /go/src/github.com/zhangkesheng/registrator \
 && dep ensure \
 && go build


FROM alpine:3.6

ENV WEAVE_VERSION=2.0.1

ADD /script/startup-script.sh /start

COPY --from=build /go/src/github.com/zhangkesheng/registrator/registrator /

WORKDIR /

CMD ["/registrator"]