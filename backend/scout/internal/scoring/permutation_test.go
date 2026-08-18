package scoring

import "testing"

func TestValidatePermutationValid(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 2}, {EngineerID: 2, Rank: 1}, {EngineerID: 3, Rank: 3}}
	active := []int{1, 2, 3}

	if err := ValidatePermutation(entries, active); err != nil {
		t.Fatalf("expected valid permutation, got error: %v", err)
	}
}

func TestValidatePermutationRejectsTie(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 2, Rank: 1}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for duplicate rank (tie)")
	}
}

func TestValidatePermutationRejectsGap(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 2, Rank: 3}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for gap in ranks (1, 3 with only 2 engineers)")
	}
}

func TestValidatePermutationRejectsMissingEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error when submission omits an active engineer")
	}
}

func TestValidatePermutationRejectsUnknownEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 99, Rank: 2}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for an engineer not in the active roster")
	}
}

func TestValidatePermutationRejectsDuplicateEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 1, Rank: 2}}
	active := []int{1}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for duplicate engineer entry")
	}
}
