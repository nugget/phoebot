.PHONY:	gazelle phoebot run deploy

deps: go-mc gazelle

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazel run :gazelle

go-mc:
	go get github.com/Tnze/go-mc@master
	go mod download
	go mod vendor

phoebot: gazelle
	clear
	bazel build :phoebot

run: gazelle phoebot
	clear
	bazel run :phoebot

mojang: gazelle
	clear
	bazel run //cmd/mojangtest

mapper: gazelle
	clear
	bazel run //cmd/mapper

deploy:
	bazel run :deploy.apply
