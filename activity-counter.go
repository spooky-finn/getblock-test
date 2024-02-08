package main

import (
	"sort"
)

type ActivityCounter struct {
	data map[string]int
}

type ActiveResult struct {
	Address       string
	ActivityCount int
}

func NewActivityCounter() *ActivityCounter {
	return &ActivityCounter{
		data: make(map[string]int),
	}
}

func (a *ActivityCounter) AddTx(tx Transaction) {
	a.data[tx.To]++
	a.data[tx.From]++
}

func (a *ActivityCounter) GetActivity() map[string]int {
	return a.data
}

func (a *ActivityCounter) GetMostMostActive(limit int) []ActiveResult {
	if limit > len(a.data) {
		limit = len(a.data)
	}

	keys := make([]string, 0, len(a.data))
	for key := range a.data {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return a.data[keys[i]] > a.data[keys[j]]
	})

	res := make([]ActiveResult, 0, limit)
	for i := 0; i < limit; i++ {
		res = append(res, ActiveResult{
			Address:       keys[i],
			ActivityCount: a.data[keys[i]],
		})
	}

	return res
}
