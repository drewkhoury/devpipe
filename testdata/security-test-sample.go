package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// This file contains intentional security issues for testing security scanners
// DO NOT use these patterns in production code!

// Issue 1: Command injection with user input (data flow)
// CodeQL will detect: go/command-injection
func handleCommand(w http.ResponseWriter, r *http.Request) {
	userInput := r.URL.Query().Get("cmd")
	cmd := exec.Command("sh", "-c", userInput) // Command injection!
	output, _ := cmd.CombinedOutput()
	w.Write(output)
}

// Issue 2: Path traversal vulnerability (data flow)
// CodeQL will detect: go/path-injection
func handleFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	data, _ := os.ReadFile(filename) // Path traversal!
	w.Write(data)
}

// Issue 3: Weak cryptographic hash with user data
// CodeQL will detect: go/weak-crypto
func hashPassword(password string) string {
	hash := md5.Sum([]byte(password)) // Weak crypto!
	return fmt.Sprintf("%x", hash)
}

func main() {
	http.HandleFunc("/cmd", handleCommand)
	http.HandleFunc("/file", handleFile)
	http.ListenAndServe(":8080", nil)
}
