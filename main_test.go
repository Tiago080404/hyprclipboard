package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestReadText(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "history.json")
	json.NewEncoder(tmpFile).Encode([]ClipItem{})

	historyFile = tmpFile.Name()
	readText("testnr1")
	data, _ := os.ReadFile(historyFile)
	var history []ClipItem
	json.Unmarshal(data, &history)

	if history[0].FullText != "testnr1" {
		t.Errorf("erwartet 'testnr1', bekommen '%s'", history[0].FullText)
	}
	defer os.Remove(tmpFile.Name())
}

func TestHistoryLength(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "history.json")
	json.NewEncoder(tmpFile).Encode([]ClipItem{})

	historyFile = tmpFile.Name()

	var history []ClipItem
	for i := range 21 {
		history = append(history, ClipItem{
			FullText: fmt.Sprintf("Eintrag %d", i),
			Display:  fmt.Sprintf("Eintrag %d", i),
			MimeType: "text/plain",
		})
	}
	data, _ := json.Marshal(history)
	os.WriteFile(tmpFile.Name(), data, 0644)

	readText("21ter Eintrag")

	result, _ := os.ReadFile(tmpFile.Name())

	var resultHistory []ClipItem
	json.Unmarshal(result, &resultHistory)
	if len(resultHistory) != 20 {
		t.Errorf("erwartet 20 Einträge, bekommen %d", len(resultHistory))
	}
	defer os.Remove(tmpFile.Name())

}
func TestDeduplication(t *testing.T) {
	history := []ClipItem{
		{FullText: "A", MimeType: "text/plain"},
		{FullText: "B", MimeType: "text/plain"},
		{FullText: "C", MimeType: "text/uri-list"},
	}

	history = deduplicateHistoryEntry(history, "B", "text/plain")
	history = deduplicateHistoryEntry(history, "C", "text/uri-list")

	if len(history) != 1 {
		t.Fatalf("expected 2 items after deduplication, got %d", len(history))
	}

	for _, item := range history {
		if item.FullText == "B" {
			t.Fatalf("duplicate 'B' was not removed")
		}
	}
}
