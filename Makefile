$(eval VERSION=$(shell git describe --always --tags --abbrev=1))

.PHONY:	gazelle phoebot run deploy datapacks clean deployall phoenixcraft ashecraft legacy nuggethaus

deps: go-mc modules gazelle

clean:
	@echo Cleaning build artifacts and output files
	@rm output/*

gazelle:
	@echo Running gazelle to process BUILD.bazelisk files for Go
	bazelisk run :gazelle -- update-repos -from_file=go.mod --prune=true
	bazelisk run :gazelle

go-mc:
	go get github.com/Tnze/go-mc@master
	go mod download

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
	kubectx nuggethaus
	stern --all-namespaces phoebot

cptest: gazelle
	clear
	bazelisk run //cmd/cptest

datapacks: clean
	@cd datapacks/phoenixcraft_postoffice &&  zip -qr ../../output/phoenixcraft-postoffice-$(VERSION).zip *
	@cd datapacks/phoenixcraft_elytra_crafting &&  zip -qr ../../output/phoenixcraft-elytra-$(VERSION).zip *
	@echo "Datapacks in output directory:"
	@ls -la output/*.zip

deployall: nuggethaus phoenixcraft ashecraft legacy
	cd db && sqitch deploy dev

nuggethaus:
	cd db && sqitch deploy prod
	bazelisk run :dev_deploy.apply

phoenixcraft:
	cd db && sqitch deploy prod
	bazelisk run :main_deploy.apply

ashecraft:
	cd db && sqitch deploy smp
	bazelisk run :smp_deploy.apply

legacy:
	cd db && sqitch deploy legacy
	bazelisk run :legacy_deploy.apply

activestate:
	cd db && sqitch deploy active
	bazelisk run :active_deploy.apply
	
dbpb:
	psql ${DATABASE_URI}

# brew cask install mysql-shell
dbcp:
	mysqlsh --sql --uri ${COREPROTECT_URI}
