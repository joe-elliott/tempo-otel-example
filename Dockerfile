FROM alpine:3.12

RUN addgroup -g 1000 app && \
    adduser -u 1000 -h /app -G app -S app
WORKDIR /app
USER app

COPY ./app .

CMD ["./app"] 