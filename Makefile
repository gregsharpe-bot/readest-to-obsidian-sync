.PHONY: test vet build docker-build helm-lint helm-template

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...

docker-build:
	docker build --tag readest-obsidian-sync:0.1.0 .

helm-lint:
	helm lint charts/readest-obsidian-sync

helm-template:
	helm template test charts/readest-obsidian-sync \
		--set aws.region=eu-west-2 \
		--set aws.sqsQueueUrl=https://example.invalid/queue \
		--set aws.s3Bucket=example
