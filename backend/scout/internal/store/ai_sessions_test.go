package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAISessionStoreCreateAndUpdateTranscript(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attributes", "main_attributes", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	sessionStore := NewAISessionStore(db)

	main, _ := mainStore.Create("test_main_ai", "Test Main AI")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	session, err := sessionStore.Create(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if string(session.Transcript) != "[]" {
		t.Fatalf("expected new session to start with empty transcript array, got %s", session.Transcript)
	}
	if session.ProposedRanking != nil {
		t.Fatalf("expected new session to have nil proposed_ranking, got %v", session.ProposedRanking)
	}

	transcript := json.RawMessage(`[{"role":"user","content":"who stood out this cycle?"}]`)
	proposed := json.RawMessage(`{"ranking":[{"engineer_id":1,"rank":1}]}`)
	if err := sessionStore.UpdateTranscript(session.ID, transcript, proposed); err != nil {
		t.Fatalf("update transcript failed: %v", err)
	}

	updated, err := sessionStore.GetByID(session.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}

	// Compare unmarshaled JSON objects to account for PostgreSQL JSON normalization
	var expectedTranscript, actualTranscript []interface{}
	if err := json.Unmarshal(transcript, &expectedTranscript); err != nil {
		t.Fatalf("unmarshal expected transcript failed: %v", err)
	}
	if err := json.Unmarshal(updated.Transcript, &actualTranscript); err != nil {
		t.Fatalf("unmarshal actual transcript failed: %v", err)
	}
	if len(actualTranscript) != len(expectedTranscript) || len(actualTranscript) == 0 {
		t.Fatalf("transcript mismatch: expected %v, got %v", expectedTranscript, actualTranscript)
	}

	var expectedProposed, actualProposed map[string]interface{}
	if err := json.Unmarshal(proposed, &expectedProposed); err != nil {
		t.Fatalf("unmarshal expected proposed failed: %v", err)
	}
	if err := json.Unmarshal(updated.ProposedRanking, &actualProposed); err != nil {
		t.Fatalf("unmarshal actual proposed failed: %v", err)
	}
	if len(actualProposed) == 0 {
		t.Fatalf("expected proposed ranking to be set, got %v", actualProposed)
	}
}

func TestAISessionStoreGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"ai_ranking_sessions"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	sessionStore := NewAISessionStore(db)

	session, err := sessionStore.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID for missing id returned error: %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil for missing session, got %+v", session)
	}
}

func TestAISessionStoreJSONRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attributes", "main_attributes", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	sessionStore := NewAISessionStore(db)

	main, _ := mainStore.Create("test_main_roundtrip", "Test Main Roundtrip")
	sub, _ := subStore.Create(main.ID, "Communication", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	session, err := sessionStore.Create(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Create a realistic, deeply nested transcript with multiple turns
	transcriptData := []map[string]interface{}{
		{
			"role":    "user",
			"content": "who stood out this cycle?",
		},
		{
			"role":    "assistant",
			"content": "Based on the metrics, Engineer A and Engineer B showed strong performance.",
		},
		{
			"role":    "user",
			"content": "can you be more specific?",
		},
		{
			"role":    "assistant",
			"content": "Engineer A: 95 score. Engineer B: 92 score.",
		},
	}
	transcriptBytes, err := json.Marshal(transcriptData)
	if err != nil {
		t.Fatalf("failed to marshal transcript: %v", err)
	}
	transcript := json.RawMessage(transcriptBytes)

	// Create a realistic, deeply nested proposed ranking with multiple engineers and scores
	proposedData := map[string]interface{}{
		"ranking": []map[string]interface{}{
			{
				"engineer_id": 1,
				"name":        "Engineer A",
				"rank":        1,
				"score":       95.5,
				"justification": map[string]interface{}{
					"main_points": []string{"strong ownership", "mentorship"},
					"confidence":  0.95,
				},
			},
			{
				"engineer_id": 2,
				"name":        "Engineer B",
				"rank":        2,
				"score":       92.3,
				"justification": map[string]interface{}{
					"main_points": []string{"technical excellence", "reliability"},
					"confidence":  0.90,
				},
			},
			{
				"engineer_id": 3,
				"name":        "Engineer C",
				"rank":        3,
				"score":       88.1,
				"justification": map[string]interface{}{
					"main_points": []string{"collaboration"},
					"confidence":  0.85,
				},
			},
		},
		"metadata": map[string]interface{}{
			"cycle_id":         cycle.ID,
			"sub_attribute_id": sub.ID,
			"generated_at":     time.Now().Unix(),
			"model":            "claude-3-sonnet",
		},
	}
	proposedBytes, err := json.Marshal(proposedData)
	if err != nil {
		t.Fatalf("failed to marshal proposed ranking: %v", err)
	}
	proposed := json.RawMessage(proposedBytes)

	// Update the session with both transcript and proposed ranking
	if err := sessionStore.UpdateTranscript(session.ID, transcript, proposed); err != nil {
		t.Fatalf("update transcript failed: %v", err)
	}

	// Retrieve and verify the exact data round-trips
	updated, err := sessionStore.GetByID(session.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}

	// Unmarshal and re-verify the transcript
	var retrievedTranscript []map[string]interface{}
	if err := json.Unmarshal(updated.Transcript, &retrievedTranscript); err != nil {
		t.Fatalf("failed to unmarshal retrieved transcript: %v", err)
	}
	if len(retrievedTranscript) != 4 {
		t.Fatalf("expected 4 transcript messages, got %d", len(retrievedTranscript))
	}
	if retrievedTranscript[0]["role"] != "user" {
		t.Fatalf("expected first message role to be 'user', got %v", retrievedTranscript[0]["role"])
	}
	if retrievedTranscript[0]["content"] != "who stood out this cycle?" {
		t.Fatalf("expected first message content to be 'who stood out this cycle?', got %v", retrievedTranscript[0]["content"])
	}

	// Unmarshal and re-verify the proposed ranking
	var retrievedProposed map[string]interface{}
	if err := json.Unmarshal(updated.ProposedRanking, &retrievedProposed); err != nil {
		t.Fatalf("failed to unmarshal retrieved proposed ranking: %v", err)
	}
	ranking, ok := retrievedProposed["ranking"].([]interface{})
	if !ok {
		t.Fatalf("expected ranking to be an array, got %T", retrievedProposed["ranking"])
	}
	if len(ranking) != 3 {
		t.Fatalf("expected 3 engineers in ranking, got %d", len(ranking))
	}

	// Verify first engineer details
	firstEngineer, ok := ranking[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first engineer to be a map, got %T", ranking[0])
	}
	if firstEngineer["name"] != "Engineer A" {
		t.Fatalf("expected first engineer name to be 'Engineer A', got %v", firstEngineer["name"])
	}
	if firstEngineer["rank"].(float64) != 1 {
		t.Fatalf("expected first engineer rank to be 1, got %v", firstEngineer["rank"])
	}
	if firstEngineer["score"].(float64) != 95.5 {
		t.Fatalf("expected first engineer score to be 95.5, got %v", firstEngineer["score"])
	}

	// Verify nested justification in first engineer
	justification, ok := firstEngineer["justification"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected justification to be a map, got %T", firstEngineer["justification"])
	}
	mainPoints, ok := justification["main_points"].([]interface{})
	if !ok {
		t.Fatalf("expected main_points to be an array, got %T", justification["main_points"])
	}
	if len(mainPoints) != 2 {
		t.Fatalf("expected 2 main points, got %d", len(mainPoints))
	}
	if mainPoints[0] != "strong ownership" {
		t.Fatalf("expected first main point to be 'strong ownership', got %v", mainPoints[0])
	}

	// Verify metadata
	metadata, ok := retrievedProposed["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata to be a map, got %T", retrievedProposed["metadata"])
	}
	if metadata["model"] != "claude-3-sonnet" {
		t.Fatalf("expected model to be 'claude-3-sonnet', got %v", metadata["model"])
	}
}

func TestAISessionStoreCreateMultipleSessions(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attributes", "main_attributes", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	sessionStore := NewAISessionStore(db)

	main, _ := mainStore.Create("test_main_multi", "Test Main Multi")
	sub1, _ := subStore.Create(main.ID, "Ownership", nil)
	sub2, _ := subStore.Create(main.ID, "Communication", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Create two sessions for different sub-attributes
	session1, err := sessionStore.Create(cycle.ID, sub1.ID)
	if err != nil {
		t.Fatalf("create session 1 failed: %v", err)
	}

	session2, err := sessionStore.Create(cycle.ID, sub2.ID)
	if err != nil {
		t.Fatalf("create session 2 failed: %v", err)
	}

	// Verify they are different sessions
	if session1.ID == session2.ID {
		t.Fatalf("expected different session IDs, got %d and %d", session1.ID, session2.ID)
	}
	if session1.SubAttributeID == session2.SubAttributeID {
		t.Fatalf("expected different sub_attribute_ids, got %d for both", session1.SubAttributeID)
	}

	// Update them independently
	transcript1 := json.RawMessage(`[{"role":"user","content":"sub 1 question"}]`)
	proposed1 := json.RawMessage(`{"ranking":[]}`)
	if err := sessionStore.UpdateTranscript(session1.ID, transcript1, proposed1); err != nil {
		t.Fatalf("update session 1 failed: %v", err)
	}

	transcript2 := json.RawMessage(`[{"role":"user","content":"sub 2 question"}]`)
	proposed2 := json.RawMessage(`{"ranking":[]}`)
	if err := sessionStore.UpdateTranscript(session2.ID, transcript2, proposed2); err != nil {
		t.Fatalf("update session 2 failed: %v", err)
	}

	// Verify they maintain separate state
	updated1, _ := sessionStore.GetByID(session1.ID)
	updated2, _ := sessionStore.GetByID(session2.ID)

	var t1Expected, t1Actual []interface{}
	json.Unmarshal(transcript1, &t1Expected)
	json.Unmarshal(updated1.Transcript, &t1Actual)
	if len(t1Actual) != len(t1Expected) || len(t1Actual) == 0 {
		t.Fatalf("session 1 transcript mismatch: expected %v, got %v", t1Expected, t1Actual)
	}

	var t2Expected, t2Actual []interface{}
	json.Unmarshal(transcript2, &t2Expected)
	json.Unmarshal(updated2.Transcript, &t2Actual)
	if len(t2Actual) != len(t2Expected) || len(t2Actual) == 0 {
		t.Fatalf("session 2 transcript mismatch: expected %v, got %v", t2Expected, t2Actual)
	}
}
