FROM golang:1.13.1-alpine3.10

RUN apk --no-cache add git

WORKDIR /go/src/app
COPY . .

# -race and therefore CGO needs gcc, we don't want it to have in our build
RUN CGO_ENABLED=0 go build -v -o tck_shoppingcart ./tck/cmd/tck_shoppingcart
RUN go install -v ./...

# multistage – copy over the binary
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=0 /go/bin/tck_shoppingcart .

EXPOSE 8080
ENV HOST 0.0.0.0
ENV PORT 8080

CMD ["./tck_shoppingcart"]