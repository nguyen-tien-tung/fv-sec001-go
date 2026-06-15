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
		if command -v powershell.exe >/dev/null 2>&1; then \
			CPU=$$(powershell.exe -NoProfile -Command "(Get-CimInstance Win32_Processor | Select-Object -First 1).Name" | tr -d '\r'); \
			RAM_BYTES=$$(powershell.exe -NoProfile -Command "(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory" | tr -d '\r'); \
			RAM=$$(awk -v b="$$RAM_BYTES" 'BEGIN { printf "%.2f GiB", b/1024/1024/1024 }'); \
			DISK_MEDIA=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty MediaType" | tr -d '\r'); \
DISK_BUS=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty BusType" | tr -d '\r'); \
DISK_NAME=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty FriendlyName" | tr -d '\r'); \
DISK_TYPE="$$DISK_MEDIA $$DISK_BUS $$DISK_NAME"; \
			INPUT_FILE_SIZE=$$(wc -c < "$(INPUT)" | awk '{ printf "%.2f MiB", $$1/1024/1024 }'); \
			echo "cpu=$$CPU"; \
			echo "ram=$$RAM"; \
			echo "disk_type=$$DISK_TYPE"; \
			echo "input_file_size=$$INPUT_FILE_SIZE"; \
		elif [ "$$(uname -s)" = "Darwin" ]; then \
			CPU=$$(sysctl -n machdep.cpu.brand_string 2>/dev/null || echo unknown); \
			RAM=$$(sysctl -n hw.memsize 2>/dev/null | awk '{ printf "%.2f GiB", $$1/1024/1024/1024 }'); \
			INPUT_FILE_SIZE=$$(wc -c < "$(INPUT)" | awk '{ printf "%.2f MiB", $$1/1024/1024 }'); \
			DISK_DEVICE=$$(df "$(INPUT)" | awk 'NR==2 {print $$1}'); \
			DISK_TYPE=$$(diskutil info "$$DISK_DEVICE" 2>/dev/null | awk -F: '/Solid State|Protocol/ { gsub(/^[ \t]+/, "", $$2); printf "%s ", $$2 } END { print "" }'); \
			if [ -z "$$DISK_TYPE" ]; then DISK_TYPE="unknown"; fi; \
			echo "cpu=$$CPU"; \
			echo "ram=$$RAM"; \
			echo "disk_type=$$DISK_TYPE"; \
			echo "input_file_size=$$INPUT_FILE_SIZE"; \
		else \
			CPU=$$(awk -F: '/model name/ { gsub(/^[ \t]+/, "", $$2); print $$2; exit }' /proc/cpuinfo 2>/dev/null || echo unknown); \
			RAM=$$(awk '/MemTotal/ { printf "%.2f GiB", $$2/1024/1024 }' /proc/meminfo 2>/dev/null || echo unknown); \
			INPUT_FILE_SIZE=$$(wc -c < "$(INPUT)" | awk '{ printf "%.2f MiB", $$1/1024/1024 }'); \
			DISK_SRC=$$(df -P "$(INPUT)" | awk 'NR==2 {print $$1}'); \
			DISK_BASE=$$(basename "$$DISK_SRC" | sed 's/[0-9]*$$//' | sed 's/p$$//'); \
			DISK_TYPE=$$(lsblk -ndo ROTA,TRAN,MODEL "/dev/$$DISK_BASE" 2>/dev/null | awk '{ rota=$$1; $$1=""; gsub(/^[ \t]+/, ""); if (rota=="0") print "SSD/NVMe " $$0; else if (rota=="1") print "HDD " $$0; else print $$0 }'); \
			if [ -z "$$DISK_TYPE" ]; then DISK_TYPE="unknown"; fi; \
			echo "cpu=$$CPU"; \
			echo "ram=$$RAM"; \
			echo "disk_type=$$DISK_TYPE"; \
			echo "input_file_size=$$INPUT_FILE_SIZE"; \
		fi; \
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

docker-bench: docker-build
	@if [ -z "$(INPUT)" ]; then \
		echo "error: INPUT is required. Usage: make docker-bench INPUT=ad_data.csv" >&2; \
		exit 2; \
	fi
	@mkdir -p benchmark "$(OUTPUT)"
	@{ \
		echo "benchmark_date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"; \
		echo "go_version=docker image"; \
		echo "os=$$(uname -s)"; \
		echo "machine=$$(uname -m)"; \
		if command -v powershell.exe >/dev/null 2>&1; then \
			CPU=$$(powershell.exe -NoProfile -Command "(Get-CimInstance Win32_Processor | Select-Object -First 1).Name" | tr -d '\r'); \
			RAM_BYTES=$$(powershell.exe -NoProfile -Command "(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory" | tr -d '\r'); \
			RAM=$$(awk -v b="$$RAM_BYTES" 'BEGIN { printf "%.2f GiB", b/1024/1024/1024 }'); \
			DISK_MEDIA=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty MediaType" | tr -d '\r'); \
			DISK_BUS=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty BusType" | tr -d '\r'); \
			DISK_NAME=$$(powershell.exe -NoProfile -Command "Get-PhysicalDisk | Select-Object -First 1 -ExpandProperty FriendlyName" | tr -d '\r'); \
			DISK_TYPE="$$DISK_MEDIA $$DISK_BUS $$DISK_NAME"; \
			INPUT_FILE_SIZE=$$(wc -c < "$(INPUT)" | awk '{ printf "%.2f MiB", $$1/1024/1024 }'); \
			echo "cpu=$$CPU"; \
			echo "ram=$$RAM"; \
			echo "disk_type=$$DISK_TYPE"; \
			echo "input_file_size=$$INPUT_FILE_SIZE"; \
		else \
			CPU=$$(awk -F: '/model name/ { gsub(/^[ \t]+/, "", $$2); print $$2; exit }' /proc/cpuinfo 2>/dev/null || echo unknown); \
			RAM=$$(awk '/MemTotal/ { printf "%.2f GiB", $$2/1024/1024 }' /proc/meminfo 2>/dev/null || echo unknown); \
			INPUT_FILE_SIZE=$$(wc -c < "$(INPUT)" | awk '{ printf "%.2f MiB", $$1/1024/1024 }'); \
			DISK_SRC=$$(df -P "$(INPUT)" | awk 'NR==2 {print $$1}'); \
			DISK_BASE=$$(basename "$$DISK_SRC" | sed 's/[0-9]*$$//' | sed 's/p$$//'); \
			DISK_TYPE=$$(lsblk -ndo ROTA,TRAN,MODEL "/dev/$$DISK_BASE" 2>/dev/null | awk '{ rota=$$1; $$1=""; gsub(/^[ \t]+/, ""); if (rota=="0") print "SSD/NVMe " $$0; else if (rota=="1") print "HDD " $$0; else print $$0 }'); \
			if [ -z "$$DISK_TYPE" ]; then DISK_TYPE="unknown"; fi; \
			echo "cpu=$$CPU"; \
			echo "ram=$$RAM"; \
			echo "disk_type=$$DISK_TYPE"; \
			echo "input_file_size=$$INPUT_FILE_SIZE"; \
		fi; \
		echo "input=$(INPUT)"; \
		echo "output=$(OUTPUT)"; \
		echo "command=docker run --rm -v \$$PWD:/data --user \$$\(id -u\):\$$\(id -g\) $(IMAGE_NAME) --input /data/$(INPUT) --output /data/$(OUTPUT)"; \
		echo "---"; \
	} > "$(BENCH_LOG)"
	@{ \
		if /usr/bin/time -v true >/dev/null 2>&1; then \
			echo "timer=/usr/bin/time -v"; \
			MSYS_NO_PATHCONV=1 /usr/bin/time -v docker run --rm \
				-v "$$PWD:/data" \
				--user "$$(id -u):$$(id -g)" \
				$(IMAGE_NAME) \
				--input "/data/$(INPUT)" \
				--output "/data/$(OUTPUT)"; \
		elif /usr/bin/time -l true >/dev/null 2>&1; then \
			echo "timer=/usr/bin/time -l"; \
			MSYS_NO_PATHCONV=1 /usr/bin/time -l docker run --rm \
				-v "$$PWD:/data" \
				--user "$$(id -u):$$(id -g)" \
				$(IMAGE_NAME) \
				--input "/data/$(INPUT)" \
				--output "/data/$(OUTPUT)"; \
		else \
			echo "timer=shell time (peak memory unavailable)"; \
			export MSYS_NO_PATHCONV=1; \
			{ time docker run --rm \
				-v "$$PWD:/data" \
				--user "$$(id -u):$$(id -g)" \
				$(IMAGE_NAME) \
				--input "/data/$(INPUT)" \
				--output "/data/$(OUTPUT)"; \
			}; \
		fi; \
	} 2>&1 | tee -a "$(BENCH_LOG)"
	@echo "docker benchmark log written to $(BENCH_LOG)"