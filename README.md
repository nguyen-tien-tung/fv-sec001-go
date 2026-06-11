# FLINTERS FV-SEC001 Ad Performance Aggregator

A Go CLI implementation for the FLINTERS Vietnam FV-SEC001 Software Engineer Challenge. The program reads an ad performance CSV, aggregates metrics by `campaign_id`, and writes the top campaigns by CTR and CPA.

## Problem summary

Input CSV schema:

```csv
campaign_id,date,impressions,clicks,spend,conversions
```

For each campaign, the CLI computes:

- `total_impressions`
- `total_clicks`
- `total_spend`
- `total_conversions`
- `CTR = total_clicks / total_impressions`
- `CPA = total_spend / total_conversions`

It generates:

- `top10_ctr.csv`: top 10 campaigns by highest CTR
- `top10_cpa.csv`: top 10 campaigns by lowest CPA, excluding campaigns with zero conversions

## Approach

The application uses a single-pass aggregation pipeline:

1. Open the input CSV.
2. Validate the header.
3. Read one CSV record at a time with `encoding/csv.Reader`.
4. Validate and parse each row.
5. Update one in-memory aggregate per campaign.
6. Sort campaign aggregates for CTR and CPA rankings.
7. Write deterministic CSV output files.

Business logic is kept in `internal` packages. The CLI entrypoint in `cmd/aggregator/main.go` only handles flags, orchestration, and process exit behavior.

## Why streaming is used

The challenge input can be around 1GB, so loading the whole CSV into memory would be unnecessary and risky. This implementation streams the file row by row and stores only campaign-level aggregate statistics.

Memory usage scales with the number of unique campaigns, not with the number of CSV rows.

## Complexity

Let:

- `n` = number of CSV data rows
- `c` = number of unique campaigns

Complexity:

```text
Time:   O(n + c log c)
Memory: O(c)
```

The `O(n)` term comes from streaming and aggregating rows. The `O(c log c)` term comes from sorting campaign aggregates for ranking. The program does not store raw CSV rows.

## Setup

Requirements:

- Go 1.22 or newer
- Make, optional but recommended
- Docker, optional

Install dependencies:

```bash
go mod download
```

This project uses only the Go standard library.

## Run command

Run with Go directly:

```bash
go run ./cmd/aggregator --input ./ad_data.csv --output ./results
```

Or use Make:

```bash
make run INPUT=ad_data.csv
```

Optional strict mode:

```bash
go run ./cmd/aggregator --input ./ad_data.csv --output ./results --strict
```

CLI flags:

| Flag | Default | Description |
|---|---:|---|
| `--input` | required | Input CSV path |
| `--output` | `results` | Output directory |
| `--strict` | `false` | Fail on the first malformed data row |

## Test command

```bash
make test
```

Equivalent command:

```bash
go test ./...
```

## Docker command

Build the image:

```bash
make docker-build
```

Show CLI help:

```bash
docker run --rm fv-sec001-aggregator
```

Run with the current directory mounted at `/data`:

```bash
docker run --rm \
  -v "$PWD:/data" \
  --user "$(id -u):$(id -g)" \
  fv-sec001-aggregator \
  --input /data/ad_data.csv \
  --output /data/results
```

Or use Make:

```bash
make docker-run INPUT=/data/ad_data.csv
```

The Docker image uses a multi-stage build and a small final `scratch` image containing the statically linked `aggregator` binary. Mount input and output paths into the container when processing local files.

## Output files

The output directory is created if it does not exist. Output files are written via temporary files and renamed into place after successful writes.

Generated files:

```text
results/top10_ctr.csv
results/top10_cpa.csv
```

Output columns are exactly:

```csv
campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA
```

Formatting:

- `total_spend`: 2 decimal places
- `CTR`: 4 decimal places
- `CPA`: 2 decimal places
- unavailable CPA: empty value

## Error handling behavior

Header errors always fail the run. The expected header is:

```csv
campaign_id,date,impressions,clicks,spend,conversions
```

Data row validation:

- `campaign_id` must be non-empty
- `impressions`, `clicks`, and `conversions` must be non-negative integers
- `spend` must be a non-negative finite number
- rows must contain the expected number of fields
- the `date` column must be present but is not semantically validated because aggregation is by campaign only

Default mode:

- malformed data rows are skipped
- invalid rows are counted
- processing continues while at least one valid row exists
- the first invalid-row samples are printed to stderr for debugging
- the run fails if no valid data rows are processed

Strict mode:

- the first malformed data row fails the run
- no successful result is reported

After a successful run, a summary is printed to stderr:

```text
processed_rows=<rows> valid_rows=<valid> invalid_rows=<invalid> campaigns=<unique_campaigns>
```

When malformed rows are encountered, up to 10 invalid-row samples are also printed to stderr. The program does not log every invalid row to avoid excessive output for large files.

## Assumptions

- Malformed rows are skipped in default mode, but the run still fails if zero valid rows remain.
- Strict mode fails fast on the first malformed data row.
- `spend` is parsed and aggregated as a `float64`; output formatting rounds `total_spend` to 2 decimal places.
- `CTR` is `0` when `total_impressions = 0`.
- `CPA` is unavailable when `total_conversions = 0`.
- Campaigns with unavailable CPA are excluded from `top10_cpa.csv`.
- Campaigns with unavailable CPA are still eligible for `top10_ctr.csv`; the CPA field is empty in that output.
- Ranking ties are deterministic and use `campaign_id` ascending.
- CTR ranking: `CTR` descending, then `campaign_id` ascending.
- CPA ranking: `CPA` ascending, then `campaign_id` ascending.

## Benchmark

Run an end-to-end benchmark:

```bash
make bench INPUT=ad_data.csv
```

The benchmark target writes output to:

```text
benchmark/benchmark.log
```

It attempts to capture wall-clock time and peak memory:

- Linux with GNU time: `/usr/bin/time -v`
- macOS/BSD time: `/usr/bin/time -l`
- Other environments: shell `time`; peak memory may be unavailable

Replace the placeholders below after running locally:

| Item | Value |
|---|---|
| Go version | TODO |
| OS / architecture | TODO |
| CPU | TODO |
| RAM | TODO |
| Disk type | TODO |
| Input file size | TODO |
| Input row count | TODO |
| Unique campaigns | TODO |
| Valid rows | TODO |
| Invalid rows | TODO |
| Wall-clock time | TODO |
| Peak memory / max RSS | TODO |

No benchmark numbers are claimed here because they depend on local hardware, disk speed, Go version, and input data shape.

## AI assistant usage

AI assistance was used during design and implementation. See [`PROMPTS.md`](./PROMPTS.md) for the prompt history and notes.
