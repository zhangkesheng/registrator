FROM golang as build

COPY . /go/src/github.com/zhangkesheng/registrator

ENV CGO_ENABLED=0

RUN go get -u github.com/golang/dep/cmd/dep

RUN cd /go/src/github.com/zhangkesheng/registrator \
 && dep ensure \
 && go build


FROM alpine:3.6

ENV WEAVE_VERSION=2.0.1

ADD /script/startup-script.sh /start

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
   apk add --no-cache tzdata wget && \
   rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
   wget -qO - http://git.oschina.net/bestmike007/files/raw/master/setenv.sh | sh -s && \
   apk add --no-cache iptables curl && \
   curl -sSL -o /usr/local/bin/weave https://github.com/weaveworks/weave/releases/download/v${WEAVE_VERSION}/weave && \
   chmod a+x /usr/local/bin/weave /start && \
   apk update && \
   apk add --no-cache docker

COPY --from=build /go/src/github.com/zhangkesheng/registrator/registrator /

WORKDIR /

CMD ["/registrator"]