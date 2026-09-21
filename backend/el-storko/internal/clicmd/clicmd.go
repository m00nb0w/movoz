package clicmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"el-storko/internal/models"
)

type ListOptions struct {
	Source string
	Status string
	EpicID *int64
}

func Add(baseURL, title string, epicID *int64, description string) (*models.WorkItem, error) {
	body := map[string]any{
		"type":        "task",
		"title":       title,
		"description": description,
	}
	if epicID != nil {
		body["parent_id"] = *epicID
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/api/work-items", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("could not reach el-storko API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiError(resp)
	}

	var item models.WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func List(baseURL string, opts ListOptions) ([]models.WorkItem, error) {
	q := url.Values{}
	if opts.Source != "" {
		q.Set("source", opts.Source)
	}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.EpicID != nil {
		q.Set("parent_id", fmt.Sprintf("%d", *opts.EpicID))
	}

	reqURL := baseURL + "/api/work-items"
	if encoded := q.Encode(); encoded != "" {
		reqURL += "?" + encoded
	}

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("could not reach el-storko API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var parsed struct {
		Items []models.WorkItem `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Items, nil
}

func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("el-storko API returned %d: %s", resp.StatusCode, string(body))
}
