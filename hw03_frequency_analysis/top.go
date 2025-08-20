package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	if text == "" {
		return nil
	}

	// Разделение текста на слова по пробельным символам
	words := strings.Fields(text)

	// Подсчет частоты каждого слова
	freq := make(map[string]int)
	for _, word := range words {
		freq[word]++
	}

	// "Сбор" слов в срез для сортировки
	type wordCount struct {
		word  string
		count int
	}
	wc := make([]wordCount, 0, len(freq))
	for w, c := range freq {
		wc = append(wc, wordCount{w, c})
	}

	// Сортировка слов: сначала по убыванию частоты, потом лексикографически
	sort.Slice(wc, func(i, j int) bool {
		if wc[i].count != wc[j].count {
			return wc[i].count > wc[j].count
		}
		return wc[i].word < wc[j].word
	})

	top := make([]string, 0, 10)
	for i, w := range wc {
		if i >= 10 {
			break
		}
		top = append(top, w.word)
	}

	return top
}
