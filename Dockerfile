FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o swarm-batch-exporter .

FROM scratch
COPY --from=builder /app/swarm-batch-exporter /swarm-batch-exporter
EXPOSE 1640
ENTRYPOINT ["/swarm-batch-exporter"]
