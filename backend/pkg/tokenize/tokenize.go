package tokenize

import (
	"os"
	"strings"

	"github.com/veer66/mapkha"
)

type Tokenizer struct {
	wordCutTH *mapkha.Wordcut
}

func (tokenize *Tokenizer) loadDictTH() (*mapkha.Dict, error) {

	var dict *mapkha.Dict
	var err error

	dictPath, exists := os.LookupEnv("DICT_PATH_TH")
	if !exists {
		dict, err = mapkha.LoadDefaultDict()
	} else {
		dict, err = mapkha.LoadDict(dictPath)
	}

	if err != nil {
		return nil, err
	}

	return dict, nil
}

func (tokenize *Tokenizer) GetTokenizerTH() (*mapkha.Wordcut, error) {

	if tokenize.wordCutTH == nil {
		dict, err := tokenize.loadDictTH()
		if err != nil {
			return nil, err
		}
		tokenize.wordCutTH = mapkha.NewWordcut(dict)

	}

	return tokenize.wordCutTH, nil
}

func (tokenize *Tokenizer) SegmentWordTH(text string) ([]string, error) {
	trimmedText := strings.Trim(text, " ")
	splitBySpace := strings.Split(trimmedText, " ")

	segmentWord := []string{}
	tokenizer, err := tokenize.GetTokenizerTH()

	if err != nil {
		return []string{}, err
	}

	for _, term := range splitBySpace {
		if len(term) > 0 {
			segmentWord = append(segmentWord, tokenizer.Segment(term)...)
		}
	}

	filteredSegmentWord := []string{}
	for _, word := range segmentWord {
		if word != "" {
			filteredSegmentWord = append(filteredSegmentWord, word)
		}
	}

	return filteredSegmentWord, nil
}
