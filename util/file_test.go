package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLineCount(t *testing.T) {

	file1 := "./data/file_0.txt"
	count, err := FileLineCount(file1)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	file2 := "./data/file_10.txt"
	count, err = FileLineCount(file2)
	assert.Nil(t, err)
	assert.Equal(t, 10, count)

	file3 := "./data/file_100.txt"
	count, err = FileLineCount(file3)
	assert.Nil(t, err)
	assert.Equal(t, 100, count)
}
