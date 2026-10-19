package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	arguments := os.Args[1:]
	switch arguments[0] {
	case "say":
		fmt.Fprint(os.Stdout, arguments[1])
		fmt.Fprint(os.Stderr, arguments[2])
		code, _ := strconv.Atoi(arguments[3])
		os.Exit(code)
	case "write":
		os.WriteFile(arguments[1], []byte(arguments[2]), 0o644)
	case "copy":
		content, _ := os.ReadFile(arguments[1])
		os.WriteFile(arguments[2], content, 0o644)
	case "stdin":
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString(0)
		fmt.Print(text)
	case "arguments":
		encoded, _ := json.Marshal(arguments[1:])
		fmt.Println(string(encoded))
	case "folder":
		directory, _ := os.Getwd()
		fmt.Print(directory)
	case "sleep":
		time.Sleep(time.Minute)
	case "stdio":
		answer()
	}
}

func answer() {
	scanner := bufio.NewScanner(os.Stdin)
	count := 0
	for scanner.Scan() {
		count++
		var request map[string]any
		json.Unmarshal(scanner.Bytes(), &request)
		if request["command"] == "garble" {
			fmt.Println("not json")
			continue
		}
		if request["command"] == "quit" {
			return
		}
		if request["command"] == "hang" {
			time.Sleep(time.Minute)
		}
		encoded, _ := json.Marshal(map[string]any{"ok": true, "count": count, "request": request, "problems": []any{}})
		fmt.Println(string(encoded))
	}
}
