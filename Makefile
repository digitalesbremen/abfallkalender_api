CONTAINER_NAME=bremer-abfallkalender-api
IMAGE_NAME=larmic/bremer-abfallkalender-api
HERUKO_APP_NAME=bremer-abfallkalender-api
VERSION_FILE=VERSION
VERSION=`cat $(VERSION_FILE)`
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

heruko-login:
	heroku login -i

heruko-container-login:
	heroku container:login

heruko-container-push:
	heroku container:push web --app ${HERUKO_APP_NAME}

heruko-container-deploy:
	heroku container:release web

heruko-container-logs:
	heroku logs --tail --app ${HERUKO_APP_NAME}

heruko-open-app:
	heroku open --app bremer-abfallkalender-api

parcel-update-package-lock:
	npm i --package-lock-only