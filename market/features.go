package market

import (
	"fmt"
	"math"

	"byodb"
)

// ComputeFeatureSnapshot calculates and persists a leakage-safe feature row
// using only candles at or before asOf. At least 21 observations are required.
func (r *Repository) ComputeFeatureSnapshot(symbol, interval string, asOf int64, featureSet string) (FeatureSnapshot, error) {
	if featureSet == "" {
		featureSet = DefaultFeatureSet
	}
	if asOf <= 0 {
		asOf = math.MaxInt64
	}
	candles, err := r.Candles(CandleQuery{Symbol: symbol, Interval: interval, From: 0, To: asOf, Limit: 21, Descending: true})
	if err != nil {
		return FeatureSnapshot{}, err
	}
	if len(candles) < 21 {
		return FeatureSnapshot{}, fmt.Errorf("%w: need 21 candles, have %d", ErrInsufficientHistory, len(candles))
	}
	for left, right := 0, len(candles)-1; left < right; left, right = left+1, right-1 {
		candles[left], candles[right] = candles[right], candles[left]
	}
	last := len(candles) - 1
	feature := FeatureSnapshot{
		Symbol: candles[last].Symbol, Interval: candles[last].Interval, Timestamp: candles[last].Timestamp,
		FeatureSet: featureSet, Return1PPM: ratioPPM(candles[last].AdjustedClose, candles[last-1].AdjustedClose),
		Return5PPM: ratioPPM(candles[last].AdjustedClose, candles[last-5].AdjustedClose),
		SMA5:       averagePrice(candles[last-4:]), SMA20: averagePrice(candles[last-19:]),
		Volatility20PPM: volatilityPPM(candles), RSI14PPM: rsiPPM(candles[last-14:]),
		VolumeSMA20: averageVolume(candles[last-19:]), ComputedAt: r.now().UTC().Unix(),
	}
	if _, err := r.db.Upsert(tableFeatures, featureRecord(feature)); err != nil {
		return FeatureSnapshot{}, err
	}
	r.metrics.featuresComputed.Add(1)
	return feature, nil
}

func (r *Repository) Feature(symbol, interval string, timestamp int64, featureSet string) (FeatureSnapshot, bool, error) {
	var err error
	if symbol, err = normalizeSymbol(symbol); err != nil {
		return FeatureSnapshot{}, false, err
	}
	if interval, err = validateInterval(interval); err != nil {
		return FeatureSnapshot{}, false, err
	}
	if featureSet == "" {
		featureSet = DefaultFeatureSet
	}
	rec := (&byodb.Record{}).AddString("symbol", symbol).AddString("interval", interval).AddInt64("timestamp", timestamp).AddString("feature_set", featureSet)
	ok, err := r.db.Get(tableFeatures, rec)
	if err != nil || !ok {
		return FeatureSnapshot{}, ok, err
	}
	return featureFromRecord(*rec), true, nil
}

func ratioPPM(current, previous int64) int64 {
	return int64(math.Round((float64(current)/float64(previous) - 1) * float64(RatioScale)))
}

func averagePrice(candles []Candle) int64 {
	var sum float64
	for _, candle := range candles {
		sum += float64(candle.AdjustedClose)
	}
	return int64(math.Round(sum / float64(len(candles))))
}

func averageVolume(candles []Candle) int64 {
	var sum float64
	for _, candle := range candles {
		sum += float64(candle.Volume)
	}
	return int64(math.Round(sum / float64(len(candles))))
}

func volatilityPPM(candles []Candle) int64 {
	returns := make([]float64, 0, len(candles)-1)
	var mean float64
	for i := 1; i < len(candles); i++ {
		value := math.Log(float64(candles[i].AdjustedClose) / float64(candles[i-1].AdjustedClose))
		returns = append(returns, value)
		mean += value
	}
	mean /= float64(len(returns))
	var sumSquares float64
	for _, value := range returns {
		delta := value - mean
		sumSquares += delta * delta
	}
	return int64(math.Round(math.Sqrt(sumSquares/float64(len(returns)-1)) * float64(RatioScale)))
}

func rsiPPM(candles []Candle) int64 {
	var gains, losses float64
	for i := 1; i < len(candles); i++ {
		delta := float64(candles[i].AdjustedClose - candles[i-1].AdjustedClose)
		if delta > 0 {
			gains += delta
		} else {
			losses -= delta
		}
	}
	if losses == 0 {
		if gains == 0 {
			return RatioScale / 2
		}
		return RatioScale
	}
	rsi := 1 - 1/(1+gains/losses)
	return int64(math.Round(rsi * float64(RatioScale)))
}

func featureFromRecord(r byodb.Record) FeatureSnapshot {
	return FeatureSnapshot{Symbol: r.Get("symbol").String(), Interval: r.Get("interval").String(), Timestamp: r.Get("timestamp").I64, FeatureSet: r.Get("feature_set").String(), Return1PPM: r.Get("return_1_ppm").I64, Return5PPM: r.Get("return_5_ppm").I64, SMA5: r.Get("sma_5").I64, SMA20: r.Get("sma_20").I64, Volatility20PPM: r.Get("volatility_20_ppm").I64, RSI14PPM: r.Get("rsi_14_ppm").I64, VolumeSMA20: r.Get("volume_sma_20").I64, ComputedAt: r.Get("computed_at").I64}
}
