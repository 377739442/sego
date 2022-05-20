package sego

import (
	"github.com/adamzy/cedar-go"
	"sync"
)

// Dictionary结构体实现了一个字串前缀树，一个分词可能出现在叶子节点也有可能出现在非叶节点
type Dictionary struct {
	trie           *cedar.Cedar // Cedar 前缀树
	maxTokenLength int          // 词典中最长的分词
	tokens         []Token      // 词典中所有的分词，方便遍历
	totalFrequency int64        // 词典中所有分词的频率之和
	lock           sync.Mutex
}

func NewDictionary() *Dictionary {
	return &Dictionary{trie: cedar.New()}
}

// 词典中最长的分词
func (dict *Dictionary) MaxTokenLength() int {
	return dict.maxTokenLength
}

// 词典中分词数目
func (dict *Dictionary) NumTokens() int {
	return len(dict.tokens)
}

// 词典中所有分词的频率之和
func (dict *Dictionary) TotalFrequency() int64 {
	return dict.totalFrequency
}

// 释放资源
func (dict *Dictionary) Close() {
	dict.trie = nil
	dict.maxTokenLength = 0
	dict.tokens = nil
	dict.totalFrequency = int64(0)
}

// 向词典中加入一个分词
func (dict *Dictionary) addToken(token Token) {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	bytes := textSliceToBytes(token.text)
	_, err := dict.trie.Get(bytes)
	if err == nil {
		return
	}

	dict.trie.Insert(bytes, dict.NumTokens())
	dict.tokens = append(dict.tokens, token)
	dict.totalFrequency += int64(token.frequency)
	if len(token.text) > dict.maxTokenLength {
		dict.maxTokenLength = len(token.text)
	}
}

// 向词典中删除一个分词
func (dict *Dictionary) delToken(token Token) {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	bytes := textSliceToBytes(token.text)
	subscript, err := dict.trie.Get(bytes)
	if err != nil {
		return
	}
	dict.tokens = append(dict.tokens[:subscript], dict.tokens[subscript+1:]...)
	err = dict.trie.Delete(bytes)
	if err != nil {
		return
	}
	dict.totalFrequency -= int64(token.frequency)
	if len(token.text) < dict.maxTokenLength {
		return
	}
	var textLen int
	for i := 0; i < dict.NumTokens(); i++ {
		textLen = len(dict.tokens[i].text)
		if textLen > dict.maxTokenLength {
			dict.maxTokenLength = textLen
		}
	}
}

// 在词典中查找和字元组words可以前缀匹配的所有分词
// 返回值为找到的分词数
func (dict *Dictionary) lookupTokens(words []Text, tokens []*Token) (numOfTokens int) {
	var id, value int
	var err error
	for _, word := range words {
		id, err = dict.trie.Jump(word, id)
		if err != nil {
			break
		}
		value, err = dict.trie.Value(id)
		if err == nil {
			tokens[numOfTokens] = &dict.tokens[value]
			numOfTokens++
		}
	}
	return
}
