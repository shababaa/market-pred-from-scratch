package market

import (
	"errors"
	"fmt"
	"reflect"

	"byodb"
)

const (
	tableMigrations  = "market_schema_migrations"
	tableInstruments = "market_instruments"
	tableCandles     = "market_candles"
	tableFeatures    = "market_features"
	tableModelRuns   = "market_model_runs"
	tableForecasts   = "market_forecasts"
	tableAnalyses    = "market_analyses"
	tableCheckpoints = "market_ingestion_checkpoints"
)

func tableDefinitions() []*byodb.TableDef {
	return []*byodb.TableDef{
		{Name: tableMigrations, Cols: []string{"version", "name", "applied_at"}, Types: []uint32{byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_INT64}, PKeys: 1, Indexes: [][]string{{"version"}}},
		{Name: tableInstruments, Cols: []string{"symbol", "name", "asset_type", "exchange", "currency", "active", "created_at", "updated_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64}, PKeys: 1, Indexes: [][]string{{"symbol"}, {"exchange", "symbol"}, {"asset_type", "symbol"}}},
		{Name: tableCandles, Cols: []string{"symbol", "interval", "timestamp", "open", "high", "low", "close", "adjusted_close", "volume", "source", "ingested_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_INT64}, PKeys: 3, Indexes: [][]string{{"symbol", "interval", "timestamp"}, {"timestamp", "symbol", "interval"}}},
		{Name: tableFeatures, Cols: []string{"symbol", "interval", "timestamp", "feature_set", "return_1_ppm", "return_5_ppm", "sma_5", "sma_20", "volatility_20_ppm", "rsi_14_ppm", "volume_sma_20", "computed_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64}, PKeys: 4, Indexes: [][]string{{"symbol", "interval", "timestamp", "feature_set"}, {"feature_set", "timestamp", "symbol", "interval"}}},
		{Name: tableModelRuns, Cols: []string{"run_id", "model_name", "model_version", "feature_set", "target", "horizon_seconds", "training_start", "training_end", "created_at", "parameters_json", "metrics_json", "dataset_hash", "status"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES}, PKeys: 1, Indexes: [][]string{{"run_id"}, {"model_name", "created_at"}, {"status", "created_at"}}},
		{Name: tableForecasts, Cols: []string{"run_id", "symbol", "interval", "as_of_timestamp", "horizon_seconds", "target_timestamp", "baseline_close", "predicted_close", "lower_bound", "upper_bound", "confidence_ppm", "has_actual", "actual_close", "evaluated_at", "absolute_error", "absolute_percentage_ppm", "direction_correct", "created_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_INT64}, PKeys: 5, Indexes: [][]string{{"run_id", "symbol", "interval", "as_of_timestamp", "horizon_seconds"}, {"symbol", "interval", "as_of_timestamp"}, {"target_timestamp", "symbol", "interval"}}},
		{Name: tableAnalyses, Cols: []string{"analysis_id", "symbol", "as_of_timestamp", "provider", "model", "prompt_version", "input_digest", "thesis", "sentiment_ppm", "confidence_ppm", "evidence_json", "forecast_run_id", "created_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64}, PKeys: 1, Indexes: [][]string{{"analysis_id"}, {"symbol", "as_of_timestamp"}, {"forecast_run_id", "created_at"}}},
		{Name: tableCheckpoints, Cols: []string{"source", "dataset", "symbol", "interval", "last_timestamp", "last_cursor", "rows_ingested", "updated_at"}, Types: []uint32{byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_BYTES, byodb.TYPE_INT64, byodb.TYPE_INT64}, PKeys: 4, Indexes: [][]string{{"source", "dataset", "symbol", "interval"}, {"updated_at", "source", "dataset"}}},
	}
}

func sameSchema(actual, expected *byodb.TableDef) bool {
	return actual.Name == expected.Name && actual.PKeys == expected.PKeys &&
		reflect.DeepEqual(actual.Cols, expected.Cols) && reflect.DeepEqual(actual.Types, expected.Types) &&
		reflect.DeepEqual(actual.Indexes, expected.Indexes)
}

// EnsureSchema applies the market schema as one atomic migration. Repeated
// calls are idempotent; an incompatible existing table fails safely.
func (r *Repository) EnsureSchema() error {
	return r.writeTransaction(func(tx *byodb.DBTX) error {
		for _, expected := range tableDefinitions() {
			if actual, ok := tx.Table(expected.Name); ok {
				if !sameSchema(actual, expected) {
					return fmt.Errorf("market schema mismatch for table %q", expected.Name)
				}
				continue
			}
			if err := tx.TableNew(expected); err != nil {
				return err
			}
		}
		migration := (&byodb.Record{}).AddInt64("version", SchemaVersion)
		exists, err := tx.Get(tableMigrations, migration)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		migration = (&byodb.Record{}).
			AddInt64("version", SchemaVersion).
			AddString("name", "initial market intelligence schema").
			AddInt64("applied_at", r.now().UTC().Unix())
		_, err = tx.Set(tableMigrations, *migration, byodb.MODE_INSERT_ONLY)
		return err
	})
}

func (r *Repository) CurrentSchemaVersion() (int64, error) {
	if _, ok := r.db.Table(tableMigrations); !ok {
		return 0, nil
	}
	var tx byodb.DBTX
	if err := r.db.Begin(&tx); err != nil {
		return 0, err
	}
	defer r.db.Abort(&tx)
	sc := &byodb.Scanner{}
	if err := tx.Scan(tableMigrations, sc); err != nil {
		return 0, err
	}
	var latest int64
	for sc.Valid() {
		var rec byodb.Record
		if err := sc.Deref(&rec); err != nil {
			return 0, err
		}
		if rec.Get("version").I64 > latest {
			latest = rec.Get("version").I64
		}
		sc.Next()
	}
	return latest, nil
}

func (r *Repository) writeTransaction(fn func(*byodb.DBTX) error) error {
	for attempt := 0; attempt < 4; attempt++ {
		var tx byodb.DBTX
		if err := r.db.Begin(&tx); err != nil {
			return err
		}
		if err := fn(&tx); err != nil {
			r.db.Abort(&tx)
			return err
		}
		if err := r.db.Commit(&tx); err != nil {
			if errors.Is(err, byodb.ErrConflict) {
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("market transaction: %w", byodb.ErrConflict)
}
