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
	var history []string
	if err := json.Unmarshal(file, &history); err != nil {
		panic(err)
	}
	if len(history) > 20 {
		history = history[:20]
	}

	if len(content) > 15 {
		fmt.Println("content too long")
		content = content[:15] + "..."
		fmt.Println("shorter", content)
	}

	if len(history) != 0 {

		if strings.TrimSpace(history[0]) == strings.TrimSpace(content) {
			fmt.Println("equals")
			return
		}
	}
	history = append([]string{content}, history...)

	newC, err := json.Marshal(history)

	if err != nil {
		panic(err)
	}
	os.WriteFile(historyFile, newC, 0644)
}
func reader() []string {
	var history []string
	file, err := os.ReadFile(historyFile)

	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(file, &history); err != nil {
		panic(err)
	}
	for i := range history {
		history[i] = strings.TrimSpace(history[i])
	}
	return history
}
func showList() {
	history := reader()
	if len(history) == 0 {
		fmt.Println("History leer")
		return
	}

	cmd := exec.Command("wofi", "--dmenu", "--prompt", "Clipboard:", "--insensitive")
	cmd.Stdin = strings.NewReader(strings.Join(history, "\n"))

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)

	}

	userInputSelect := out.String()
	cmdCopy := exec.Command("wl-copy")
	cmdCopy.Stdin = strings.NewReader(userInputSelect)
	cmdCopy.Run()

}

func getHistoryFilePath() string {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	filePath := filepath.Join(configDir, "cliphist", "cliphist.json")
	err := os.MkdirAll(filepath.Dir(filePath), 0750)
	if err != nil {
		panic(err)
	}

	history := []string{}
	data, _ := json.Marshal(history)
	os.WriteFile(filePath, data, 0644)
	return filePath
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
