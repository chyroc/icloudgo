package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

func setIfNotEmpty[T comparable](m map[string]T, key string, val T) map[string]T {
	var empty T
	if val != empty {
		m[key] = val
	}
	return m
}

var (
	invalidChars = []rune{' ', '!', '@', '#', '$', '%', '^', '&', '(', ')', '+', '=', '[', ']', '{', '}', ';', ':', '\'', '"', ',', '.', '<', '>', '/', '?', '\\', '|'}
	invalidChar  = map[rune]bool{}
)

func init() {
	for _, v := range invalidChars {
		invalidChar[v] = true
	}
}

func cleanName(s string) string {
	l := []rune(s)
	for i, v := range l {
		if invalidChar[v] {
			l[i] = '_'
		}
	}
	return string(l)
}

func cleanFilename(s string) string {
	ext := filepath.Ext(s)
	base := s[:len(s)-len(ext)]
	return cleanName(base) + ext
}

type set[T comparable] map[T]struct{}

func newSet[T comparable](initValue ...T) set[T] {
	res := make(map[T]struct{})
	for _, v := range initValue {
		res[v] = struct{}{}
	}
	return res
}

func (s set[T]) Add(item T) {
	s[item] = struct{}{}
}

func (s set[T]) Remove(item T) {
	if _, ok := s[item]; ok {
		delete(s, item)
	}
}

func (s set[T]) Len() int {
	return len(s)
}

func (s set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

func (s set[T]) String() string {
	var res []string
	for k := range s {
		res = append(res, fmt.Sprintf("%v", k))
	}
	return strings.Join(res, ",")
}
