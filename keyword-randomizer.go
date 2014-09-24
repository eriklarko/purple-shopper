package main

import (
	"os"
	"log"
	"bufio"
	"time"
	"math/rand"
	"strings"
)

var words []string = nil

func generateRandomSearchString() string {
	if words == nil {
		setupWordsDatabase()
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strings.ToLower(words[r.Intn(len(words))])
}

func setupWordsDatabase() {
	lines, err := ReadLines("CommonWords.txt")
	if err != nil {
		log.Println("Unable to open word database")
		log.Fatal(err)
	}

	words = lines
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
