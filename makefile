.PHONY:  default  deploy  destroy  refresh  test  test-coverage  test-docker  test-release

default: test

deploy:
	GOOS=linux GOARCH=amd64 go build -o dist/slackbot main.go
	(cd dist ; zip -9 slackbot.zip slackbot)
	cd terraform/ && terraform apply -auto-approve

destroy:
	cd terraform/ && terraform destroy

refresh:
	cookiecutter gh:sjansen/cookiecutter-golang --output-dir .. --config-file .cookiecutter.yaml --no-input --overwrite-if-exists
	git checkout go.mod go.sum

test:
	@scripts/run-all-tests
	@echo ========================================
	@git grep TODO  -- '**.go' || true
	@git grep FIXME -- '**.go' || true

test-coverage: test-docker
	go tool cover -html=dist/coverage.txt

test-docker:
	@scripts/docker-up-test

test-release:
	git stash -u -k
	goreleaser release --rm-dist --skip-publish
	-git stash pop
