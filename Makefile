build:
	go build -o swarm-batch-exporter

docker:
	docker build . -t blossomlabs/swarm-batch-prometheus
