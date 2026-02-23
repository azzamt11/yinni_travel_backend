.PHONY: init
# install dependencies
init:
	npm install

.PHONY: run
# run auth service
run:
	node ./index.js -conf ./configs

.PHONY: dev
# run in dev mode
dev:
	node --watch ./index.js -conf ./configs

.PHONY: help
help:
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@echo '  init    install dependencies'
	@echo '  run     run auth service'
	@echo '  dev     run auth service in watch mode'

.DEFAULT_GOAL := help
