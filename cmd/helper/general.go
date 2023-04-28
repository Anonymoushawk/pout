package helper

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/atotto/clipboard"
)

func WriteTextToClipboard(text string) bool {
	err := clipboard.WriteAll(text)

	return err == nil
}

// `FormatBytes` formats the passed number of bytes into a string with a byte size suffix.
func FormatBytes(n int) string {
	// Kilobyte and Megabyte sizes (in bytes).
	const (
		kilobyteSize = 1024
		megabyteSize = kilobyteSize * kilobyteSize
	)

	// Determine the suffix to use depending on number of bytes.
	switch {
	case n >= megabyteSize:
		return fmt.Sprintf("%.1fMB", float64(n)/float64(megabyteSize))
	case n >= kilobyteSize:
		return fmt.Sprintf("%.1fKB", float64(n)/float64(kilobyteSize))
	default:
		return fmt.Sprintf("%dB", n)
	}
}

// `GenerateSessionKey` generates a random string from numbers and letters.
func GenerateSessionKey(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	key := make([]byte, length)
	for i := range key {
		// Generate random character to add to the resulting key.
		key[i] = CHARS[rand.Intn(len(CHARS))]
	}

	return string(key)
}

// Selection of characters to use in the key generation process.
var CHARS = "ABCDEF123456789"
