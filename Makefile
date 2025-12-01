

.PHONY: docker run

NAME := yapi

install:
	# run make install from ./cli/Makefile
	cd cli && $(MAKE) install

web:
	docker build . -t ${NAME}:latest -f Dockerfile.webapp

web-run:
	-docker stop yapi
	-docker rm yapi
	docker run --name yapi -p 3000:3000 ${NAME}:latest


