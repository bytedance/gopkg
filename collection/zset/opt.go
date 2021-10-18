package zset

// RangeOpt describes the whether the min/max is exclusive in score range.
type RangeOpt struct {
	ExcludeMin bool
	ExcludeMax bool
}
