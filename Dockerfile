FROM golang:1.12.0-alpine
RUN apk add --no-cache git gcc g++
COPY . /app
WORKDIR /app
RUN go install .
CMD /go/bin/registry-webhook
