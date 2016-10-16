package randomkeyword

import (
	"os"
	"log"
	"bufio"
	"time"
	"math/rand"
	"strings"
)

var words []string = nil

func GenerateRandomSearchString() string {
	if words == nil {
		setupWordsDatabase()
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strings.ToLower(words[r.Intn(len(words))])
}

func setupWordsDatabase() {
	lines, error := ReadLines("src/randomkeyword/CommonWords.txt")
	if error != nil {
		log.Println("Unable to open word database")
		log.Fatal(error)
	}

	words = lines
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, error := os.Open(path)
	if error != nil {
		return nil, error
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
