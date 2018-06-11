# ----------------------------------------------------------------------------------------
# Image: Builder
# ----------------------------------------------------------------------------------------
FROM golang:alpine as builder

# setup the environment
ENV TZ=Europe/Berlin

# install dependencies
RUN apk --update --no-cache add git tzdata
WORKDIR /go/src/github.com/faryon93/whalepost
ADD ./ ./

# build the go binary
RUN go build -ldflags \
        '-X "main.BuildTime='$(date -Iminutes)'" \
         -X "main.GitCommit='$(git rev-parse --short HEAD)'" \
         -X "main.GitBranch='$(git rev-parse --abbrev-ref HEAD)'" \
         -X "main.BuildNumber='$CI_BUILDNR'" \
         -s -w' \
         -v -o /tmp/whalepost .

# ----------------------------------------------------------------------------------------
# Image: Deployment
# ----------------------------------------------------------------------------------------
FROM alpine:latest
MAINTAINER Maximilian Pachl <m@ximilian.info>

# setup the environment
ENV TZ=Europe/Berlin

RUN apk --update --no-cache add ca-certificates tzdata su-exec

# add relevant files to container
COPY --from=builder /tmp/whalepost /usr/sbin/whalepost
ADD entry.sh /entry.sh

# make binary executable
RUN chown nobody:nobody /usr/sbin/whalepost && \
    chown nobody:nobody /entry.sh && \
    chmod +x /usr/sbin/whalepost && \
    chmod +x /entry.sh

EXPOSE 8000
CMD /usr/sbin/whalepost
