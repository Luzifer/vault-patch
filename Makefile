default: install

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

install:
	go install -ldflags='-X main.version=$(shell git describe --tags || git rev-parse --short HEAD)'
