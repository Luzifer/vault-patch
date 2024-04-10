default: install

publish:
	bash ci/build.sh

install:
	go install -ldflags='-X main.version=$(shell git describe --tags || git rev-parse --short HEAD)'
