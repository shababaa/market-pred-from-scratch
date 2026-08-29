# Professional project roadmap

Each phase ends in a demo, measurable acceptance criteria, and a clean commit.
This prevents the project from becoming a collection of half-connected AI
features.

## Phase 1 — market data foundation (complete)

- Versioned domain schema and safe schema introspection
- Fixed-point OHLCV types and validation
- Atomic idempotent ingestion and CSV adapter
- Composite time-series indexes and bounded range API
- Point-in-time technical features
- Model/forecast/evaluation lineage
- Walk-forward persistence baseline and metrics
- Provider-neutral LLM interface and auditable analysis storage
- Unit, integration, race, and benchmark coverage

Exit criteria: deterministic demo runs from an empty file and every test passes.

## Phase 2 — real data pipeline

- Add one official/documented market-data provider adapter
- Rate limiting, exponential backoff, pagination, and resumable checkpoints
- Trading calendar and missing-candle detection
- Corporate-action handling and data-quality report
- Incremental feature recomputation

Exit criteria: ingest and validate at least five years of daily data for a
multi-asset universe with a repeatable CLI job.

## Phase 3 — prediction service

- Time-based train/validation/test splits
- Linear, moving-average, and gradient-boosted baselines
- Rolling-window feature generation without leakage
- Hyperparameter/config tracking through `market_model_runs`
- MAE, RMSE, MAPE, direction accuracy, calibration, and benchmark comparison
- Walk-forward and regime-segmented evaluation

Exit criteria: produce a reproducible model card and show whether each model
beats the persistence baseline after costs are excluded from the claim.

## Phase 4 — grounded LLM analyst

- Add a hosted or local LLM adapter behind `LLMClient`
- Retrieve trusted news/filings and store source metadata
- Structured output validation and retry policy
- Prompt-injection defenses and strict context separation
- Link each thesis to feature, forecast, and source IDs
- Evaluation set for factuality, citation quality, consistency, and abstention

Exit criteria: the analyst refuses unsupported claims and every visible claim
can be traced to stored evidence.

## Phase 5 — service and dashboard

- HTTP/gRPC API with timeouts, request IDs, and graceful shutdown
- Background ingestion/evaluation workers
- React dashboard for charts, features, forecasts, confidence, and model cards
- Prometheus-compatible metrics and structured logs
- Docker, CI, load tests, and reproducible demo deployment

Exit criteria: one command starts a documented end-to-end system with health
checks and a seeded demo.

## Phase 6 — database systems depth

- Inter-process file locking and read-only replicas
- Write batching/profile-driven allocation improvements
- Table dropping and efficient range deletion
- Prefix compression for time-series keys
- Fuzzing, crash-fault injection, and long-running recovery tests

Exit criteria: publish before/after benchmark evidence and document every
durability assumption.

## Resume-ready metrics to earn, not invent

- Number of candles/assets and years of history ingested
- Sustained rows/second and p50/p95 indexed query latency
- Crash/recovery and fuzz-test duration
- Model error relative to persistence baseline on an untouched test window
- LLM factuality/abstention score on a labelled evaluation set
- API load-test throughput and p95 latency

Numbers belong on a resume only after the benchmark command, data snapshot, and
evaluation method are committed so an interviewer can reproduce them.
