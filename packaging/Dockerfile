FROM alpine:3.16

RUN mkdir /launcher
COPY build-out/containerz /launcher
WORKDIR /launcher

EXPOSE 9999

CMD ["./containerz", "start"]