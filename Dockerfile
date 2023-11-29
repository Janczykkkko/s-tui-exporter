FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build .

FROM scratch 

COPY --from=0 /app/s-tui-exporter .

EXPOSE 8080

CMD ["/s-tui-exporter"]