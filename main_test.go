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
	var history []ClipItem
	for i := range 21 {
		history = append(history, ClipItem{
			FullText: fmt.Sprintf("Eintrag %d", i),
			Display:  fmt.Sprintf("Eintrag %d", i),
			MimeType: "text/plain",
		})
	}
	history = checkHistoryLength(history)

	if len(history) != 19 {
		t.Fatalf("expected 19 but got %d", len(history))
	}

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
