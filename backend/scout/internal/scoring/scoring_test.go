package scoring

import "testing"

func TestRankToScore(t *testing.T) {
	cases := []struct {
		name string
		rank int
		n    int
		want float64
	}{
		{"rank 1 of 5 is 100", 1, 5, 100},
		{"last rank of 5 is 50", 5, 5, 50},
		{"middle rank of 5", 3, 5, 75},
		{"only engineer scores 100", 1, 1, 100},
		{"rank 1 of 11 is 100", 1, 11, 100},
		{"rank 11 of 11 is 50", 11, 11, 50},
		{"rank 6 of 11 is midpoint", 6, 11, 75},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RankToScore(tc.rank, tc.n)
			if got != tc.want {
				t.Fatalf("RankToScore(%d, %d) = %v, want %v", tc.rank, tc.n, got, tc.want)
			}
		})
	}
}
