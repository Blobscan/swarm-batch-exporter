build:
	go build -o swarm-batch-exporter

docker:
	docker build . -t swarm-batch-exporter
