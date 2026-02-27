package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var historyFile string

type ClipItem struct {
	FullText string `json:"full_text"`
	Display  string `json:"display"`
}

func main() {
	historyFile = getHistoryFilePath()
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "list" {
		showList()
		return
	}
	if len(args) > 0 && args[0] == "delete" {
		deleteHistory()
		return
	}
	cmd := exec.Command("wl-paste")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Fehler")
		return
	}
	clipBoardContent := out.String()
	fmt.Println(clipBoardContent)
	readFile(clipBoardContent)
}

func readFile(content string) {
	content = strings.TrimSpace(content)
	file, err := os.ReadFile(historyFile)

	if err != nil {
		panic(err)
	}
	var history []ClipItem
	if err := json.Unmarshal(file, &history); err != nil {
		panic(err)
	}
	if len(history) != 0 {
		for index, item := range history {
			if strings.TrimSpace(item.FullText) == strings.TrimSpace(content) {
				history = append(history[:index], history[index+1:]...)
				break
			}
		}
	}

	if len(history) > 20 {
		history = history[:20]
	}

	display := content
	if len(content) > 15 {
		display = display[:15] + "..."
	}

	newItem := ClipItem{FullText: content, Display: display}
	history = append([]ClipItem{newItem}, history...)
	newC, err := json.Marshal(history)

	if err != nil {
		panic(err)
	}
	os.WriteFile(historyFile, newC, 0644)
}
func reader() []ClipItem {
	var history []ClipItem
	file, err := os.ReadFile(historyFile)

	if err != nil {
		return []ClipItem{}
	}
	if len(file) == 0 {
		return []ClipItem{}
	}

	if err := json.Unmarshal(file, &history); err != nil {
		panic(err)
	}

	return history
}
func showList() {
	history := reader()
	if len(history) == 0 {
		return
	}
	var displayList []string
	for _, item := range history {
		displayList = append(displayList, item.Display)
	}

	cmd := exec.Command("wofi", "--dmenu", "--prompt", "Clipboard:", "--insensitive")
	cmd.Stdin = strings.NewReader(strings.Join(displayList, "\n"))

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)

	}

	userInputSelect := out.String()

	for _, item := range history {
		if item.Display == userInputSelect {

			cmdCopy := exec.Command("wl-copy")
			cmdCopy.Stdin = strings.NewReader(item.FullText)
			cmdCopy.Run()
			break
		}
	}

}

func getHistoryFilePath() string {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	filePath := filepath.Join(configDir, "cliphist", "cliphist.json")
	err := os.MkdirAll(filepath.Dir(filePath), 0750)
	if err != nil {
		panic(err)
	}
	if fileExists(filePath) {
		return filePath
	}
	history := []ClipItem{{FullText: "Test", Display: "what"}}
	data, _ := json.Marshal(history)
	os.WriteFile(filePath, data, 0644)
	return filePath
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func deleteHistory() {
	history := reader()
	history = history[:1]

	newC, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}
	os.WriteFile(historyFile, newC, 0644)
}
