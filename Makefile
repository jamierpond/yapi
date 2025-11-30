

.PHONY: docker run

NAME := yapi

build:
	docker build . -t ${NAME}:latest -f Dockerfile.web

run:
	-docker stop yapi
	-docker rm yapi
	docker run --name yapi -p 3000:3000 ${NAME}:latest
