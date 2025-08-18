package hw03frequencyanalysis

import (
	"slices"
	"strings"
)

func Top10(text string) (top []string) {
	textTokens := strings.Fields(text)
	counts := make(map[string]int, len(textTokens))
	for _, w := range textTokens {
		counts[w] += 1
	}

	words := make([]string, len(counts))
	i := 0
	for k, _ := range counts {
		words[i] = k
		i++
	}

	slices.SortFunc(words, func(a, b string) int {
		aCount, aOk := counts[a]
		bCount, bOk := counts[b]
		if !aOk || !bOk {
			return 0
		}
		if aCount < bCount {
			return 1
		} else if aCount > bCount {
			return -1
		} else {
			return strings.Compare(a, b)
		}
	})

	limit := 10
	for i, word := range words {
		top = append(top, word)
		if (i + 1) >= limit {
			break
		}
	}
	return
}
