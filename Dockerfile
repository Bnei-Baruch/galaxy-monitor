ARG work_dir=/go/src/github.com/Bnei-Baruch/galaxy-monitor

FROM golang:1.14-alpine3.11 as build

LABEL maintainer="bbfsdev@gmail.com"

ARG work_dir

ENV GOOS=linux \
	CGO_ENABLED=0

RUN apk update && \
    apk add --no-cache \
    git

WORKDIR ${work_dir}
COPY . .

RUN go test $(go list ./...) \
    && go build


FROM alpine:3.11
ARG work_dir
WORKDIR /app
COPY ./misc/wait-for /wait-for
COPY --from=build ${work_dir}/galaxy-monitor .

EXPOSE 8080
CMD ["./galaxy-monitor", "server"]
