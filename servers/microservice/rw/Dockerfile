FROM alpine
COPY ./rw /questionqueue
RUN apk add --no-cache ca-certificates
EXPOSE 8000
ENTRYPOINT [ "/questionqueue" ]