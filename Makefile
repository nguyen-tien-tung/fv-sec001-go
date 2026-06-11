APP_NAME := aggregator
IMAGE_NAME := fv-sec001-aggregator
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)
OUTPUT ?= results
BENCH_LOG := benchmark/benchmark.log

.PHONY: test build run bench clean docker-build docker-run

test:
	go test ./...

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/aggregator

run:
	@if [ -z "$(INPUT)" ]; then \
		echo "error: INPUT is required. Usage: make run INPUT=ad_data.csv" >&2; \
		exit 2; \
	fi
	go run ./cmd/aggregator --input "$(INPUT)" --output "$(OUTPUT)"

bench: build
	@if [ -z "$(INPUT)" ]; then \
		echo "error: INPUT is required. Usage: make bench INPUT=ad_data.csv" >&2; \
		exit 2; \
	fi
	@mkdir -p benchmark "$(OUTPUT)"
	@{ \
		echo "benchmark_date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"; \
		echo "go_version=$$(go version)"; \
		echo "os=$$(uname -s)"; \
		echo "machine=$$(uname -m)"; \
		echo "input=$(INPUT)"; \
		echo "output=$(OUTPUT)"; \
		echo "command=$(BIN) --input $(INPUT) --output $(OUTPUT)"; \
		echo "---"; \
	} > "$(BENCH_LOG)"
	@{ \
		if /usr/bin/time -v true >/dev/null 2>&1; then \
			echo "timer=/usr/bin/time -v"; \
			/usr/bin/time -v "$(BIN)" --input "$(INPUT)" --output "$(OUTPUT)"; \
		elif /usr/bin/time -l true >/dev/null 2>&1; then \
			echo "timer=/usr/bin/time -l"; \
			/usr/bin/time -l "$(BIN)" --input "$(INPUT)" --output "$(OUTPUT)"; \
		else \
			echo "timer=shell time (peak memory unavailable)"; \
			time "$(BIN)" --input "$(INPUT)" --output "$(OUTPUT)"; \
		fi; \
	} 2>&1 | tee -a "$(BENCH_LOG)"
	@echo "benchmark log written to $(BENCH_LOG)"

clean:
	rm -rf $(BIN_DIR)
	rm -f results/top10_ctr.csv results/top10_cpa.csv
	rm -f $(BENCH_LOG)

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	@if [ -z "$(INPUT)" ]; then \
		echo "error: INPUT is required. Usage: make docker-run INPUT=/data/ad_data.csv" >&2; \
		exit 2; \
	fi
	docker run --rm -v "$$PWD:/data" --user "$$(id -u):$$(id -g)" $(IMAGE_NAME) --input "$(INPUT)" --output /data/$(OUTPUT)
