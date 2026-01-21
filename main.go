package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var historyFile = "/home/tiago/Projects/goprojects/cliphistory/cliphist.json"

func main() {
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
		fmt.Println("historz", history)
	}
	fmt.Println(len(history))
	for i := range history {
		history[i] = strings.TrimSpace(history[i])
	}

	if strings.TrimSpace(history[0]) == strings.TrimSpace(content) {
		fmt.Println("equals")
		return
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

	if len(history) == 0 {
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
	fmt.Println("Ausgewählter Clip:", userInputSelect)
	cmdCopy.Run()

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
