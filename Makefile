CONTAINER_NAME=abfallkalender_api
IMAGE_NAME=larmic/abfallkalender_api
VERSION_FILE=VERSION
VERSION=`cat $(VERSION_FILE)`
IMAGE_TAG=${VERSION}

docker-build:
	@echo "Remove docker image if already exists"
	docker rmi -f ${IMAGE_NAME}:${IMAGE_TAG}
	@echo "Build go docker image"
	DOCKER_BUILDKIT=1 docker build --build-arg VERSION=${VERSION} -t ${IMAGE_NAME}:${IMAGE_TAG} .
	@echo "Prune intermediate images"
	docker image prune --filter label=stage=intermediate -f

docker-run:
	docker run -d --rm --name ${CONTAINER_NAME} ${IMAGE_NAME}

docker-stop:
	docker stop ${CONTAINER_NAME}