# PROMPTS.md

This file contains the raw prompts used with an AI coding assistant while solving the FLINTERS Vietnam FV-SEC001 Software Engineer Challenge.

## Prompt 1 — Understand and plan

```text
I am solving the FLINTERS Vietnam FV-SEC001 Software Engineer Challenge: Ad Performance Aggregator.

Requirements:
- Build a CLI app that processes a large CSV file around 1GB.
- CSV columns: campaign_id, date, impressions, clicks, spend, conversions.
- Aggregate by campaign_id:
  - total_impressions
  - total_clicks
  - total_spend
  - total_conversions
  - CTR = total_clicks / total_impressions
  - CPA = total_spend / total_conversions
- Generate:
  - results/top10_ctr.csv: top 10 campaigns with highest CTR
  - results/top10_cpa.csv: top 10 campaigns with lowest CPA, excluding zero conversions
- Must be memory-efficient and suitable for large files.
- Include tests, error handling, README, and benchmark notes.

I want to implement this in Go.

Before writing code, propose:
1. Architecture
2. Package structure
3. Data model
4. Sorting/tie-breaking strategy
5. Error-handling strategy for malformed rows
6. Test cases
7. CLI flags
8. Benchmark approach

Important constraints:
- Do not load the whole CSV into memory.
- Use streaming CSV reading.
- Keep dependencies minimal.
- Prefer deterministic output.
- Do not write implementation yet.
```

## Prompt 2 — Challenge the design

```text
Review the proposed design critically as if you are a senior backend engineer reviewing a hiring-test submission.

Look for:
- Incorrect CTR/CPA logic
- Memory risks with a 1GB file
- Float formatting problems
- Missing edge cases
- Bad package boundaries
- Overengineering
- Underengineering
- Ambiguous behavior in ties
- CSV header handling issues
- Invalid row handling
- Whether the result files will be deterministic

Return a revised design and explicit acceptance criteria.
Do not write code yet.
```

## Prompt 3 — Create the project skeleton

```text
Implement the project skeleton in Go.

Target structure:

.
├── cmd/aggregator/main.go
├── internal/aggregator/aggregator.go
├── internal/aggregator/aggregator_test.go
├── internal/csvio/reader.go
├── internal/csvio/writer.go
├── internal/model/campaign.go
├── results/.gitkeep
├── benchmark/.gitkeep
├── Makefile
├── Dockerfile
├── README.md
└── PROMPTS.md

Rules:
- Use Go standard library only unless there is a strong reason.
- CLI flags:
  - --input required
  - --output default "results"
  - --strict default false
- In non-strict mode, skip malformed rows and count them.
- In strict mode, fail on first malformed row.
- Create output directory if missing.
- Keep main.go thin.
- Put business logic in internal packages.
- Do not implement everything in main.go.
```

## Prompt 4 — Implement streaming aggregation

```text
Implement streaming CSV aggregation.

Requirements:
- Use encoding/csv.Reader.
- Read and validate the header.
- Process one record at a time.
- Store only per-campaign aggregate stats in memory.
- Parse:
  - campaign_id as non-empty string
  - impressions as non-negative integer
  - clicks as non-negative integer
  - spend as non-negative float
  - conversions as non-negative integer
- Skip malformed records in non-strict mode and increment invalid row count.
- Return summary:
  - total rows processed
  - valid rows
  - invalid rows
  - unique campaigns
- Do not load all rows into memory.
- Handle total_impressions = 0 by setting CTR to 0.
- Handle total_conversions = 0 by setting CPA as null/unavailable internally.

Please implement clean, testable code.
```

## Prompt 5 — Implement ranking and CSV output

```text
Implement ranking and CSV output.

Output columns must be exactly:

campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA

Generate:
1. top10_ctr.csv
   - highest CTR first
   - include campaigns even if conversions = 0
   - if CPA is unavailable, output empty value for CPA

2. top10_cpa.csv
   - lowest CPA first
   - exclude campaigns with total_conversions = 0

Tie-breaking:
- For CTR ranking: CTR desc, campaign_id asc
- For CPA ranking: CPA asc, campaign_id asc

Formatting:
- total_spend: 2 decimal places
- CTR: 4 decimal places
- CPA: 2 decimal places, empty if unavailable

Write tests for sorting and formatting.
```

## Prompt 6 — Tests

```text
Add comprehensive tests.

Test cases:
1. Aggregates multiple rows for the same campaign correctly.
2. Computes CTR correctly.
3. Computes CPA correctly.
4. Handles zero conversions by excluding from top10_cpa.csv.
5. Handles zero impressions by CTR = 0.
6. Skips malformed rows in non-strict mode.
7. Fails on malformed rows in strict mode.
8. Validates deterministic tie-breaking.
9. Verifies CSV output header exactly.
10. Verifies numeric formatting:
    - spend with 2 decimals
    - CTR with 4 decimals
    - CPA with 2 decimals

Use table-driven tests where appropriate.
Do not require the 1GB file for unit tests.
```

## Prompt 7 — Run and fix

```text
Run:
- gofmt
- go test ./...
- go vet ./...

Fix all failures.

Do not change behavior to make tests pass unless the change matches the challenge requirements.
Explain any non-obvious fixes.
```

## Prompt 8 — Performance review

```text
Review the implementation for performance and memory use with a 1GB CSV.

Check for:
- accidental storage of all rows
- unnecessary string allocations
- inefficient sorting
- excessive logging per row
- poor error messages
- scanner/token-size limitations
- CSV reader reuse issues
- file handle leaks

Suggest improvements only if they materially improve correctness, memory, or readability.
Then apply safe improvements.
```

## Prompt 9 — Benchmark script

```text
Add a simple benchmark workflow.

Requirements:
- Add Makefile targets:
  - make test
  - make run INPUT=ad_data.csv
  - make bench INPUT=ad_data.csv
  - make docker-build
  - make docker-run INPUT=/data/ad_data.csv
- make bench should measure:
  - wall-clock time
  - peak memory if available on the OS
- Write benchmark output to benchmark/benchmark.log.
- Keep it portable enough for Linux/macOS, but document limitations.

Do not introduce heavy dependencies.
```

## Prompt 10 — Dockerfile

```text
Create a production-quality Dockerfile for this Go CLI.

Requirements:
- Multi-stage build.
- Final image should be small.
- Binary name: aggregator.
- Default command should show help or require --input.
- Support mounting input/output directories.

Also update README with Docker usage examples.
```

## Prompt 11 — README

```text
Write a professional README.md for the hiring challenge.

Include:
1. Problem summary
2. Approach
3. Why streaming is used
4. Complexity:
   - Time: O(n + c log c)
   - Memory: O(c)
   where n = rows and c = unique campaigns
5. Setup
6. Run command
7. Test command
8. Docker command
9. Output files
10. Error handling behavior
11. Assumptions:
    - malformed rows skipped in default mode
    - strict mode fails fast
    - CPA unavailable when conversions = 0
    - deterministic tie-breaking by campaign_id
12. Benchmark section with placeholder values that I will replace after running locally
13. AI assistant usage note pointing to PROMPTS.md

Keep it concise but complete.
Do not exaggerate performance.
Do not claim benchmark numbers unless they are provided.
```

## Prompt 12 — Final review as interviewer

```text
Act as the FLINTERS reviewer evaluating this repository.

Review for:
- Correctness
- Large-file readiness
- Code readability
- Error handling
- Tests
- README clarity
- Docker usability
- Whether PROMPTS.md demonstrates good AI-agent usage
- Any red flags that would make this look junior

Give me a prioritized punch list.
Do not rewrite code yet.
```

## Prompt 13 — Final cleanup

```text
Apply only high-confidence fixes from the punch list.

Rules:
- Do not add unnecessary abstractions.
- Do not add new dependencies.
- Keep the project easy to review.
- Preserve deterministic output.
- Ensure go test ./... passes.
- Ensure README matches actual behavior.
```
