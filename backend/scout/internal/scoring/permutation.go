package scoring

import "fmt"

type RankEntry struct {
	EngineerID int
	Rank       int
}

// ValidatePermutation enforces F6: a ranking submission must be a strict
// 1..N permutation of exactly the active roster — no ties, no gaps, no
// duplicates, no missing or extra engineers.
func ValidatePermutation(entries []RankEntry, activeEngineerIDs []int) error {
	if len(entries) != len(activeEngineerIDs) {
		return fmt.Errorf("expected exactly %d ranked engineers (the active roster), got %d", len(activeEngineerIDs), len(entries))
	}

	activeSet := make(map[int]bool, len(activeEngineerIDs))
	for _, id := range activeEngineerIDs {
		activeSet[id] = true
	}

	seenEngineers := make(map[int]bool, len(entries))
	seenRanks := make(map[int]bool, len(entries))
	for _, e := range entries {
		if !activeSet[e.EngineerID] {
			return fmt.Errorf("engineer %d is not in the active roster for this cycle", e.EngineerID)
		}
		if seenEngineers[e.EngineerID] {
			return fmt.Errorf("engineer %d appears more than once in the submission", e.EngineerID)
		}
		seenEngineers[e.EngineerID] = true

		if e.Rank < 1 || e.Rank > len(entries) {
			return fmt.Errorf("rank %d is out of range 1..%d", e.Rank, len(entries))
		}
		if seenRanks[e.Rank] {
			return fmt.Errorf("rank %d is used more than once (ties are not allowed)", e.Rank)
		}
		seenRanks[e.Rank] = true
	}

	for rank := 1; rank <= len(entries); rank++ {
		if !seenRanks[rank] {
			return fmt.Errorf("rank %d is missing (ranks must be a contiguous 1..%d sequence)", rank, len(entries))
		}
	}

	return nil
}
