CONTAINER_NAME=bremer-abfallkalender-api
IMAGE_NAME=larmic/abfallkalender_api
VERSION=local-build
IMAGE_TAG=${VERSION}

.PHONY: help run go-test go-lint docker-build docker-build-lambda docker-build-multiarch docker-run docker-stop

help:
	@echo "Available targets:"
	@echo "  run                    - run the API locally on :8080"
	@echo "  go-test                - run all tests with the race detector"
	@echo "  go-lint                - run gofmt and go vet"
	@echo "  docker-build           - build the standard image (single arch)"
	@echo "  docker-build-lambda    - build the AWS Lambda image (single arch)"
	@echo "  docker-build-multiarch - cross-compile all published architectures"
	@echo "  docker-run             - run the standard image on :8080"
	@echo "  docker-stop            - stop the running container"

run:
	go run .

go-test:
	@echo "Running go tests"
	go test -race -v ./...

go-lint:
	@echo "Checking formatting"
	@test -z "$$(gofmt -l .)" || (echo "Not gofmt-formatted:" && gofmt -l . && exit 1)
	@echo "Running go vet"
	go vet ./...

docker-build:
	@echo "Build go docker image (standard target)"
	DOCKER_BUILDKIT=1 docker build --target runner-standard --build-arg VERSION=${VERSION} -t ${IMAGE_NAME}:${IMAGE_TAG} .
	@echo "Prune intermediate images"
	docker image prune --filter label=stage=intermediate -f

docker-build-lambda:
	@echo "Build go docker image (lambda target)"
	DOCKER_BUILDKIT=1 docker build --target runner-lambda --build-arg VERSION=${VERSION} -t ${IMAGE_NAME}:${IMAGE_TAG}-lambda .
	@echo "Prune intermediate images"
	docker image prune --filter label=stage=intermediate -f

# Mirrors what CI publishes. Requires a buildx builder with the docker-container
# driver; the result stays in the build cache because a multi-arch image cannot
# be loaded into the local docker image store.
docker-build-multiarch:
	docker buildx build --target runner-standard \
		--platform=linux/amd64,linux/arm64,linux/arm/v7 \
		--build-arg VERSION=${VERSION} .

docker-run:
	docker run -p 8080:8080 --rm --name ${CONTAINER_NAME} ${IMAGE_NAME}:${IMAGE_TAG}

docker-stop:
	docker stop ${CONTAINER_NAME}
