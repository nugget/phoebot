$(eval VERSION=$(shell git describe --always --tags --abbrev=1))

.PHONY:	gazelle phoebot run deploy datapacks clean deployall phoenixcraft ashecraft legacy nuggethaus

deps: modules gazelle

clean:
	@echo Cleaning build artifacts and output files
	@rm -f output/*
	@rm -f phoebot

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazelisk run :gazelle -- update-repos -from_file=go.mod --prune=true -to_macro=deps.bzl%go_dependencies
	bazelisk run :gazelle

modules: 
	go get -u ./...
	go mod tidy
	go mod verify

phoebot: gazelle
	clear
	bazelisk build :phoebot

run: gazelle phoebot
	clear
	bazelisk run :phoebot

mojang: gazelle
	clear
	bazelisk run //cmd/mojangtest

mapper: gazelle
	clear
	bazelisk run //cmd/mapper

deploy:
	bazelisk run :deploy.apply

log:
	kubectx microk8s
	stern --all-namespaces phoebot

cptest: gazelle
	clear
	bazelisk run //cmd/cptest

datapacks: clean
	@cd datapacks/phoenixcraft_postoffice &&  zip -qr ../../output/phoenixcraft-postoffice-$(VERSION).zip *
	@cd datapacks/phoenixcraft_elytra_crafting &&  zip -qr ../../output/phoenixcraft-elytra-$(VERSION).zip *
	@echo "Datapacks in output directory:"
	@ls -la output/*.zip

deployall: nuggetcraft phoenixcraft legacy activestate

nuggetcraft:
	cd db && sqitch deploy nuggetcraft
	bazelisk run :nuggetcraft.apply

phoenixcraft:
	cd db && sqitch deploy phoenixcraft
	bazelisk run :phoenixcraft.apply

ashecraft:
	cd db && sqitch deploy ashecraft
	bazelisk run :ashecraft.apply

legacy:
	cd db && sqitch deploy legacy
	bazelisk run :legacy.apply

activestate:
	cd db && sqitch deploy activestate
	bazelisk run :activestate.apply
	
dbpb:
	psql ${DATABASE_URI}

# brew cask install mysql-shell
dbcp:
	mysqlsh --sql --uri ${COREPROTECT_URI}
