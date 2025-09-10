INSTALL_DIR := install
PYTHON_BIN := python
PRECOMMIT_FILE := .pre-commit-config.yaml
VENV := venv
REQUIREMENTS_FILE := requirements.txt
PACKAGE_LOCK := package-lock.json
COMMITLINT_FILE := commitlint.config.js
NPM_BIN := npm
GOLANGCI_LINT_BIN := golangci-lint
GOLANG_BIN := go
GORELEASER_BIN := goreleaser
TEST_EXECUTABLE_NAME := ./dist/hwttest

dev: \
	$(INSTALL_DIR)/.dir.stamp \
	$(INSTALL_DIR)/.asdf.stamp \
	$(INSTALL_DIR)/.venv.stamp \
	$(INSTALL_DIR)/.npm.stamp \
	$(INSTALL_DIR)/.precommit.stamp

$(INSTALL_DIR)/.dir.stamp:
	@mkdir -p $(INSTALL_DIR)
	@touch $@

$(INSTALL_DIR)/.asdf.stamp:
	@asdf install
	@touch $@

$(INSTALL_DIR)/.npm.stamp: $(INSTALL_DIR)/.asdf.stamp $(PACKAGE_LOCK)
	@$(NPM_BIN) install
	@touch $@

$(INSTALL_DIR)/.venv.stamp: $(REQUIREMENTS_FILE) $(INSTALL_DIR)/.asdf.stamp
	@test -d "$(VENV)" || $(PYTHON_BIN) -m venv "$(VENV)"
	. "$(VENV)/bin/activate"; \
		pip install --upgrade pip; \
		pip install -r "$(REQUIREMENTS_FILE)"
	@touch $@

$(INSTALL_DIR)/.precommit.stamp: $(PRECOMMIT_FILE) $(INSTALL_DIR)/.venv.stamp
	@. "$(VENV)/bin/activate"; pre-commit install && \
		pre-commit install --hook-type commit-msg
	@touch $@

test/unit:
	@$(GOLANGCI_LINT_BIN) config verify
	@$(GORELEASER_BIN) check
	@$(GOLANG_BIN) test ./internal/... -v

fmt:
	@$(GOLANG_BIN) mod tidy
	@env GOFUMPT_SPLIT_LONG_LINES=on $(GOLANGCI_LINT_BIN) fmt ./...

lint:
	@$(GOLANGCI_LINT_BIN) run

build/test:
	@mkdir -p ./dist/
	@$(GOLANG_BIN) build -v -o $(TEST_EXECUTABLE_NAME) main.go

test/integration: build/test
	@HWT_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v ./test/... --debug
