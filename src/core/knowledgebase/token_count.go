package knowledgebase

import (
	"strings"
	"unicode"
)

// WARNING: This is a simplified estimation of BERT token count and should be used with caution.
// Limitations:
// - Uses basic character-based heuristics instead of proper WordPiece tokenization
// - Does not account for different BERT vocabularies or model variations
// - May be inaccurate for non-English text and special characters
// - Does not implement actual BERT tokenization rules
//
// For production use or when accurate token counts are required:
// 1. Use a proper BERT tokenizer library
// 2. Consider model-specific tokenization rules
// 3. Account for different languages and character sets
//
// This implementation is intended only for rough estimation purposes.

// EstimateBertTokenCount provides a rough estimation of BERT token count
func EstimateBertTokenCount(text string) int {
	if text == "" {
		return 0
	}

	// Add [CLS] and [SEP] tokens
	count := 2

	// Remove extra whitespace
	text = strings.TrimSpace(text)
	if text == "" {
		return count
	}

	// Split into words
	words := strings.Fields(text)

	for _, word := range words {
		count += estimateWordTokens(word)
	}

	return count
}

func estimateWordTokens(word string) int {
	// Handle punctuation
	if len(word) == 1 && unicode.IsPunct(rune(word[0])) {
		return 1
	}

	// Handle numbers
	if isNumber(word) {
		return len(word) // Each numeric character might be an independent token
	}

	// General word estimation
	// BERT typically breaks long words into smaller pieces
	length := len(word)
	if length <= 4 {
		return 1
	} else {
		// Estimate that long words will be broken into multiple tokens
		return (length + 3) / 4 // Rough estimation of one token per 4 characters
	}
}

func isNumber(word string) bool {
	for _, r := range word {
		if !unicode.IsDigit(r) && r != '.' && r != ',' {
			return false
		}
	}
	return true
}
