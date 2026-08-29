# Changelog

## v0.2.0 — Market data foundation

- Added defensive database catalog introspection and operational storage stats.
- Added an atomic, idempotent versioned market schema with eight domain tables.
- Added fixed-point instruments and OHLCV candles with validation.
- Added sorted atomic batch ingestion, CSV import, checkpoints, and metrics.
- Added indexed ascending and descending time-series range queries.
- Added point-in-time returns, moving averages, volatility, RSI, and volume features.
- Added model-run lineage, immutable forecast identity, realized evaluation, and
  a leakage-safe persistence baseline.
- Added a provider-neutral LLM client boundary, typed response validation,
  structured point-in-time context, quantitative-forecast grounding, and input
  digests.
- Added the `marketdb` CLI, deterministic end-to-end demo, tests, benchmarks,
  architecture notes, and delivery roadmap.

## v0.1.0 — Textbook database

- Copy-on-write B+ tree with fixed 4 KiB pages.
- Durable page store, checksummed alternating metadata, and crash fallback.
- Versioned free-page reuse, transactions, snapshots, and optimistic conflicts.
- Relational records, composite keys, range scans, and secondary indexes.
- Simplified SQL-like parser/interpreter and interactive CLI.
