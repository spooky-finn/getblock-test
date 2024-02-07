package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSignature(t *testing.T) {
	getBlockAPI := NewGetBlockAPI()
	analyzer := NewByteCodeAnalyser(getBlockAPI)

	method := "setCount"
	params := []interface{}{"uint32"}
	expected := "4ff3eaa4"
	actual := analyzer.contractMetodSignature(method, params)

	assert.Equal(t, expected, actual)
}
