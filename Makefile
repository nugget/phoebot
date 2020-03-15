$(eval VERSION=$(shell git describe --always --tags --abbrev=1))

.PHONY:	gazelle phoebot run deploy datapacks clean deployall phoenixcraft ashecraft legacy nuggethaus

deps: go-mc modules gazelle

clean:
	@echo Cleaning build artifacts and output files
	@rm output/*

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazel run :gazelle -- update-repos -from_file=go.mod --prune=true
	bazel run :gazelle

go-mc:
	go get github.com/Tnze/go-mc@master
	go mod download

modules: 
	go get -u ./...
	go mod tidy
	go mod verify

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

log:
	kubectx nuggethaus
	stern --all-namespaces phoebot

cptest: gazelle
	clear
	bazel run //cmd/cptest

datapacks: clean
	@cd datapacks/phoenixcraft_postoffice &&  zip -qr ../../output/phoenixcraft-postoffice-$(VERSION).zip *
	@cd datapacks/phoenixcraft_elytra_crafting &&  zip -qr ../../output/phoenixcraft-elytra-$(VERSION).zip *
	@echo "Datapacks in output directory:"
	@ls -la output/*.zip

deployall: phoenixcraft ashecraft legacy
	cd db && sqitch deploy dev

nuggethaus:
	cd db && sqitch deploy prod
	bazel run :main_deploy.apply

phoenixcraft:
	cd db && sqitch deploy prod
	bazel run :main_deploy.apply

ashecraft:
	cd db && sqitch deploy smp
	bazel run :smp_deploy.apply

legacy:
	cd db && sqitch deploy legacy
	bazel run :legacy_deploy.apply
	
dbpb:
	psql ${DATABASE_URI}

# brew cask install mysql-shell
dbcp:
	mysqlsh --sql --uri ${COREPROTECT_URI}
