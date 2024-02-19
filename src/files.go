package main

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

func NameToInt(name string) int {
	base := name
	i := strings.LastIndex(name, ".")
	if i != -1 {
		base = name[:i]
	}
	v, e := strconv.Atoi(base)
	if e != nil {
		return 0
	}
	return v
}

type ByName []string

func (s ByName) Len() int {
	return len(s)
}
func (s ByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByName) Less(i, j int) bool {
	a, b := s[i], s[j]
	return NameToInt(a) < NameToInt(b)
}

// 获取已有的文件名称切片
func ListNames(dir string) ([]string, error) {
	if _, e := os.Stat(dir); os.IsNotExist(e) {
		return make([]string, 0), nil
	}
	es, e := os.ReadDir(dir)
	if e != nil {
		return make([]string, 0), e
	}
	names := make([]string, 0, len(es))
	for _, e := range es {
		names = append(names, e.Name())
	}
	sort.Sort(ByName(names))
	return names, nil
}
