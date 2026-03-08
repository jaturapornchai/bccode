package dataimport

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

// ========== Trie Data Structure ==========

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

type Trie struct {
	root       *TrieNode
	maxWordLen int
	words      []string // เก็บคำทั้งหมดสำหรับ spell check
}

func NewTrie() *Trie {
	return &Trie{
		root:  &TrieNode{children: make(map[rune]*TrieNode)},
		words: []string{},
	}
}

func (t *Trie) Insert(word string) {
	node := t.root
	runes := []rune(word)

	for _, r := range runes {
		if node.children[r] == nil {
			node.children[r] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[r]
	}

	node.isEnd = true
	t.words = append(t.words, word) // เก็บคำไว้สำหรับ spell check

	if len(runes) > t.maxWordLen {
		t.maxWordLen = len(runes)
	}
}

func (t *Trie) FindLongest(runes []rune, start int) (string, int) {
	node := t.root
	lastMatch := ""
	lastLength := 0

	for i := start; i < len(runes) && i < start+t.maxWordLen; i++ {
		r := runes[i]
		if node.children[r] == nil {
			break
		}
		node = node.children[r]

		if node.isEnd {
			lastMatch = string(runes[start : i+1])
			lastLength = i - start + 1
		}
	}

	return lastMatch, lastLength
}

// IsInDict ตรวจสอบว่าคำอยู่ใน dictionary หรือไม่
func (t *Trie) IsInDict(word string) bool {
	node := t.root
	runes := []rune(word)

	for _, r := range runes {
		if node.children[r] == nil {
			return false
		}
		node = node.children[r]
	}

	return node.isEnd
}

// ========== Spell Checker ==========

// LevenshteinDistance คำนวณระยะห่างระหว่าง 2 คำ
func LevenshteinDistance(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) == 0 {
		return len(r2)
	}
	if len(r2) == 0 {
		return len(r1)
	}

	matrix := make([][]int, len(r1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(r2)+1)
	}

	for i := 0; i <= len(r1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(r2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(r1); i++ {
		for j := 1; j <= len(r2); j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(r1)][len(r2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ========== Tokenizer ==========

type Tokenizer struct {
	trie            *Trie
	dictSize        int
	autoCorrect     bool
	maxEditDistance int
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		trie:            NewTrie(),
		autoCorrect:     true, // เปิดใช้งาน auto correct by default
		maxEditDistance: 2,    // ยอมรับความผิดพลาด 1-2 ตัวอักษร
	}
}

// EnableAutoCorrect เปิด/ปิดการแก้ไขคำอัตโนมัติ
func (t *Tokenizer) EnableAutoCorrect(enable bool) {
	t.autoCorrect = enable
}

// SetMaxEditDistance กำหนดระยะความผิดพลาดสูงสุดที่ยอมรับ
func (t *Tokenizer) SetMaxEditDistance(distance int) {
	t.maxEditDistance = distance
}

// LoadDict โหลด Thai dictionary
func (t *Tokenizer) LoadDict(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			t.trie.Insert(word)
			count++
		}
	}

	t.dictSize = count
	return scanner.Err()
}

// DictSize returns number of words loaded
func (t *Tokenizer) DictSize() int {
	return t.dictSize
}

// TokenizeWithSpaces - tokenize and keep whitespace/punctuation
func (t *Tokenizer) TokenizeWithSpaces(text string) []string {
	runes := []rune(text)
	result := []string{}
	i := 0

	for i < len(runes) {
		r := runes[i]

		// ภาษาไทย
		if t.isThai(r) {
			tokens := t.tokenizeThai(runes, &i)
			result = append(result, tokens...)
			continue
		}

		// ภาษาอังกฤษ
		if t.isEnglish(r) {
			word := t.collectEnglish(runes, &i)
			result = append(result, word)
			continue
		}

		// ตัวเลข
		if unicode.IsDigit(r) {
			num := t.collectDigits(runes, &i)
			result = append(result, num)
			continue
		}

		// เก็บช่องว่างและเครื่องหมาย (แทนที่จะข้าม)
		result = append(result, string(r))
		i++
	}

	return result
}

// Tokenize - ตัดคำ return เฉพาะคำ ไม่มีเครื่องหมาย
func (t *Tokenizer) Tokenize(text string) []string {
	runes := []rune(text)
	result := []string{}
	i := 0

	for i < len(runes) {
		r := runes[i]

		// ภาษาไทย
		if t.isThai(r) {
			tokens := t.tokenizeThai(runes, &i)
			result = append(result, tokens...)
			continue
		}

		// ภาษาอังกฤษ (ไม่แก้ไขคำผิด เพราะส่วนมากเป็นทับศัพท์)
		if t.isEnglish(r) {
			word := t.collectEnglish(runes, &i)
			result = append(result, word)
			continue
		}

		// ตัวเลข
		if unicode.IsDigit(r) {
			num := t.collectDigits(runes, &i)
			result = append(result, num)
			continue
		}

		// ข้ามช่องว่างและเครื่องหมายทั้งหมด
		i++
	}

	return result
}

func (t *Tokenizer) isThai(r rune) bool {
	return r >= 0x0E00 && r <= 0x0E7F
}

func (t *Tokenizer) isEnglish(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func (t *Tokenizer) tokenizeThai(runes []rune, i *int) []string {
	result := []string{}
	start := *i

	for *i < len(runes) && t.isThai(runes[*i]) {
		word, length := t.trie.FindLongest(runes, *i)

		if length > 0 {
			result = append(result, word)
			*i += length
		} else {
			// ไม่เจอในพจนานุกรม
			unknownWord := t.collectUnknownThai(runes, i)

			// ถ้าเปิด auto correct และคำยาวพอ ให้ลองแก้ไข
			if t.autoCorrect && len([]rune(unknownWord)) > 1 {
				corrected := t.findClosestWord(unknownWord)
				result = append(result, corrected)
			} else {
				result = append(result, unknownWord)
			}
		}

		// ป้องกันไม่ให้ค้าง
		if *i == start {
			*i++
			break
		}
		start = *i
	}

	return result
}

// collectUnknownThai รวบรวมคำไทยที่ไม่อยู่ใน dict
func (t *Tokenizer) collectUnknownThai(runes []rune, i *int) string {
	start := *i
	maxLen := 15 // จำกัดความยาวคำ

	for *i < len(runes) && t.isThai(runes[*i]) && (*i-start) < maxLen {
		// ลองหาคำถัดไปที่ตรงกับ dict
		_, length := t.trie.FindLongest(runes, *i)
		if length > 0 {
			break
		}
		*i++
	}

	// ถ้าไม่เจออะไรเลย เอาอย่างน้อย 1 ตัว
	if *i == start {
		*i++
	}

	return string(runes[start:*i])
}

// findClosestWord หาคำที่ใกล้เคียงที่สุดใน dict (สำหรับภาษาไทยเท่านั้น)
func (t *Tokenizer) findClosestWord(word string) string {
	if len(t.trie.words) == 0 {
		return word
	}

	// ถ้าคำอยู่ใน dict แล้ว ไม่ต้องแก้
	if t.trie.IsInDict(word) {
		return word
	}

	bestMatch := word
	bestDistance := t.maxEditDistance + 1
	wordRunes := []rune(word)

	// หาคำที่ใกล้เคียงที่สุด
	for _, dictWord := range t.trie.words {
		dictRunes := []rune(dictWord)

		// ข้ามถ้าความยาวต่างกันมากเกินไป
		if abs(len(wordRunes)-len(dictRunes)) > t.maxEditDistance {
			continue
		}

		distance := LevenshteinDistance(word, dictWord)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = dictWord
		}
	}

	// ถ้าเจอคำที่ใกล้เคียงพอ ใช้คำนั้น ไม่งั้นใช้คำเดิม
	if bestDistance <= t.maxEditDistance {
		return bestMatch
	}

	return word
}

func (t *Tokenizer) collectEnglish(runes []rune, i *int) string {
	start := *i
	for *i < len(runes) && t.isEnglish(runes[*i]) {
		*i++
	}
	return strings.ToLower(string(runes[start:*i]))
}

func (t *Tokenizer) collectDigits(runes []rune, i *int) string {
	start := *i
	for *i < len(runes) && unicode.IsDigit(runes[*i]) {
		*i++
	}
	return string(runes[start:*i])
}
