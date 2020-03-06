.PHONY:	gazelle phoebot run deploy clean release datapacks

deps: go-mc modules gazelle

clean:
	rm -rf output

gazelle:
	@echo Running gazelle to process BUILD.bazel files for Go
	bazel run :gazelle -- update-repos -from_file=go.mod
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

datapacks:
	mkdir -p output
	cd datapacks && zip -r ../output/datapack-elytra-crafting.zip phoenixcraft_elytra_crafting

release: datapacks

log:
	kubectx nuggethaus
	stern --all-namespaces phoebot

cptest: gazelle
	clear
	bazel run //cmd/cptest

dbpb:
	psql ${DATABASE_URI}

# brew cask install mysql-shell
dbcp:
	mysqlsh --sql --uri ${COREPROTECT_URI}
