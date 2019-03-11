FROM alpine
COPY ./gateway /questionqueue
RUN apk add --no-cache ca-certificates
EXPOSE 8123
ENTRYPOINT [ "/questionqueue" ]