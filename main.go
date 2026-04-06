package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var historyFile string

const tmpDir = "/tmp/cliphist-img-wofi"

type ClipItem struct {
	FullText string `json:"full_text"`
	Display  string `json:"display"`
	ImageB64 string `json:"image_b64"`
	MimeType string `json:"mime_type"`
	ImgPath  string `json:"img_path"`
}

func main() {
	historyFile = getHistoryFilePath()

	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		fmt.Println("dir exits")
	} else {
		err := os.Mkdir(tmpDir, 0750)

		if err != nil {
			fmt.Println("Could not create dir")
			return
		}
	}

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "list" {
		showList()
		return
	}
	if len(args) > 0 && args[0] == "delete" {
		deleteHistory()
		return
	}
	cmd := exec.Command("wl-paste", "--list-types")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Could not run cmd")
		return
	}
	clipBoardContent := strings.TrimSpace(out.String())
	if strings.Contains(clipBoardContent, "image/png") {
		readImageFile("image/png")
	} else if strings.Contains(clipBoardContent, "x-special/gnome-copied-files") {
		readFile("text/uri-list")
	} else {
		fmt.Println("normal text")
		cmd := exec.Command("wl-paste")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Println("Could not read content of type text")
			return
		}
		content := strings.TrimSpace(out.String())
		readText(content)
	}
}

func readImageFile(mime string) {
	cmd := exec.Command("wl-paste", "-t", mime)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Print("Couldnt read imae")
		return
	}
	imageData := out.Bytes()
	data := base64.StdEncoding.EncodeToString(imageData)

	imgPath := addImageForWofiDisplay(imageData)
	display := strings.Join([]string{"img:", imgPath}, "")
	file, err := os.ReadFile(historyFile)

	if err != nil {
		panic(err)
	}
	var history []ClipItem
	if err := json.Unmarshal(file, &history); err != nil {
		panic(err)
	}

	if len(history) > 20 {
		history = history[:20]
	}

	if len(history) != 0 {
		for index, item := range history {
			if strings.TrimSpace(item.ImageB64) == strings.TrimSpace(data) {
				history = append(history[:index], history[index+1:]...)
				break
			}
		}
	}
	newItem := ClipItem{Display: display, ImageB64: data, MimeType: mime, ImgPath: imgPath}
	history = append([]ClipItem{newItem}, history...)
	newC, err := json.Marshal(history)

	if err != nil {
		panic(err)
	}
	os.WriteFile(historyFile, newC, 0644)
}

func readText(content string) {
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

	newItem := ClipItem{FullText: content, Display: strings.TrimSpace(display), MimeType: "text/plain"}
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

	cmd := exec.Command("wofi", "--dmenu", "--prompt", "Clipboard:", "--insensitive", "--allow-images")
	cmd.Stdin = strings.NewReader(strings.Join(displayList, "\n"))

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)

	}

	userInputSelect := strings.TrimSpace(out.String())

	for _, item := range history {

		if item.Display == userInputSelect {

			if item.MimeType == "text/plain" {

				cmdCopy := exec.Command("wl-copy")
				cmdCopy.Stdin = strings.NewReader(item.FullText)
				cmdCopy.Run()
				break
			} else if item.MimeType == "text/uri-list" {
				cmdCopy := exec.Command("wl-copy", "-t", item.MimeType)
				cmdCopy.Stdin = strings.NewReader(item.FullText)
				cmdCopy.Run()
				break
			} else {
				data, _ := base64.StdEncoding.DecodeString(item.ImageB64)
				cmdCopy := exec.Command("wl-copy", "-t", item.MimeType)
				cmdCopy.Stdin = bytes.NewReader(data)
				cmdCopy.Run()
				break
			}
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

func addImageForWofiDisplay(data []byte) string {
	hashedValue := sha256.Sum256(data)
	name := hex.EncodeToString(hashedValue[:]) + ".png"
	fileNamePath := filepath.Join(tmpDir, name)

	os.WriteFile(fileNamePath, data, 0644)
	return fileNamePath
}

func deleteHistory() {
	history := []ClipItem{}

	data, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}
	os.WriteFile(historyFile, data, 0644)
}

func readFile(uriList string) {
	cmd := exec.Command("wl-paste", "-t", uriList)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return
	}
	content := strings.TrimSpace(out.String())

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

	display := strings.Join([]string{"📄", content}, "")

	newItem := ClipItem{FullText: content, Display: display, MimeType: uriList}
	history = append([]ClipItem{newItem}, history...)
	newC, err := json.Marshal(history)

	if err != nil {
		panic(err)
	}

	os.WriteFile(historyFile, newC, 0644)
}
