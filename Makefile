.PHONY:	gazelle phoebot run deploy

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazel run :gazelle

phoebot: gazelle
	clear
	bazel build :phoebot

run: gazelle phoebot
	clear
	bazel run :phoebot

mojang: gazelle
	clear
	bazel run //cmd/mojangtest

deploy:
	bazel run :deploy.apply
