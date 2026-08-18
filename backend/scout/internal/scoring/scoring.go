package scoring

// RankToScore converts a 1..N rank into a 50-100 score via linear
// interpolation: rank 1 -> 100, rank N -> 50, evenly spaced (F7).
// N == 1 is a special case (no interpolation possible) and scores 100.
func RankToScore(rank, n int) float64 {
	if n <= 1 {
		return 100
	}
	return 100 - float64(rank-1)*50/float64(n-1)
}
