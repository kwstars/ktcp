

.PHONY: lint
# lint
lint:
	docker run --rm -v "${PWD}":/app -w /app golangci/golangci-lint:latest \
	sh -c "GOPROXY=https://goproxy.cn,direct GO111MODULE=on golangci-lint run"

.PHONY: lint-fix
# lint-fix
lint-fix:
	docker run --rm -v "${PWD}":/app -w /app golangci/golangci-lint:latest \
	sh -c "GOPROXY=https://goproxy.cn,direct GO111MODULE=on golangci-lint run --concurrency=2 --fix --timeout 1m"