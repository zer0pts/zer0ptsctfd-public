FROM alpine:3.11

RUN apk add --update --no-cache ca-certificates tzdata
COPY ./scoreserver /scoreserver

ENTRYPOINT ["/scoreserver"]
