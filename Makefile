.PHONY:	gazelle phoebot run deploy

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazel run :gazelle

phoebot: gazelle
	clear
	bazel build //cmd/phoebot

run: gazelle phoebot
	clear
	bazel run //cmd/phoebot

deploy:
	bazel run //cmd/phoebot:deploy.apply
