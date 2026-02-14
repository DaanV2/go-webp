package histograms

const NUM_BUCKETS = 256

type HistogramBuckets [NUM_BUCKETS]uint32

// Keeping track of histograms, indexed by HistoIx.
// Ideally, this would just be a struct with meaningful fields, but the
// calculation of `entropy_comp` uses the index. One refactoring at a time :)
type Histograms struct {
	// TODO: Jesus it a large struct
	Category [HistoTotal]HistogramBuckets
}
