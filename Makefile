CONTAINER_NAME=bremer-abfallkalender-api
IMAGE_NAME=larmic/bremer-abfallkalender-api
HERUKO_APP_NAME=bremer-abfallkalender-api
VERSION_FILE=VERSION
VERSION=local-build
IMAGE_TAG=${VERSION}

go-test:
	@echo "Running go tests"
	go test -v ./...

docker-build:
	@echo "Remove docker image if already exists"
	docker rmi -f ${IMAGE_NAME}:${IMAGE_TAG}
	@echo "Build go docker image"
	DOCKER_BUILDKIT=1 docker build --build-arg VERSION=${VERSION} -t ${IMAGE_NAME}:${IMAGE_TAG} .
	@echo "Prune intermediate images"
	docker image prune --filter label=stage=intermediate -f

docker-run:
	docker run -p 8080:8080 --rm --name ${CONTAINER_NAME} ${IMAGE_NAME}:${VERSION}

docker-stop:
	docker stop ${CONTAINER_NAME}

parcel-update-package-lock:
	npm i --package-lock-only