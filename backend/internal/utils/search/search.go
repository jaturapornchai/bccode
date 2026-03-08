package search

import (
	"strings"

	m "github.com/veer66/mapkha"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var wordCut *m.Wordcut

func GetTokenizer() (*m.Wordcut, error) {

	if wordCut == nil {
		dict, err := m.LoadDict("./tdict-std.txt")
		if err != nil {
			return nil, err
		}

		wordCut = m.NewWordcut(dict)
	}

	return wordCut, nil
}

func CreateTextFilter(searchFields []string, query string) []interface{} {
	searchTerms := ExtractSearchTerms(query)
	fieldFilters := GenerateFieldFilters(searchFields, searchTerms)

	return fieldFilters
}

func ExtractSearchTerms(query string) []string {
	trimmedQuery := strings.Trim(query, " ")
	splitBySpace := strings.Split(trimmedQuery, " ")

	searchTerms := []string{}
	tokenizer, err := GetTokenizer()

	if err != nil {
		return []string{query}
	}

	for _, term := range splitBySpace {
		if len(term) > 0 {
			searchTerms = append(searchTerms, tokenizer.Segment(term)...)
		}
	}

	return searchTerms
}

func GenerateFieldFilters(searchFields []string, searchTerms []string) []interface{} {
	fieldFilters := []interface{}{}

	for _, field := range searchFields {
		termFilters := []interface{}{}

		for _, searchTerm := range searchTerms {
			termFilters = append(termFilters, bson.M{
				field: primitive.Regex{
					Pattern: searchTerm,
					Options: "i",
				},
			})
		}

		if len(termFilters) > 0 {
			fieldFilters = append(fieldFilters, bson.M{"$and": termFilters})
		}

	}

	return fieldFilters
}
