FROM alpine
COPY ./gateway /questionqueue
RUN apk add --no-cache ca-certificates
EXPOSE 80
ENTRYPOINT [ "/questionqueue" ]