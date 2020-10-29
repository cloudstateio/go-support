FROM golang:1.14.4-alpine3.12 as builder

RUN apk --no-cache add git
RUN apk --no-cache add ca-certificates

WORKDIR /go/src/app
COPY . .
#
# -race and therefore CGO needs gcc, we don't want it to have in our build
RUN CGO_ENABLED=0 go build -v -o tck_eventsourced ./tck/cmd/tck_eventsourced
RUN go install -v ./...
#
# multistage – copy over the binary
FROM alpine:latest
RUN mkdir -p /srv/
WORKDIR /srv
COPY --from=builder /go/bin/tck_eventsourced .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENV HOST 0.0.0.0
ENV PORT 8080
CMD ["./tck_eventsourced"]
