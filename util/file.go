package util

import (
	"bufio"
	"os"
)

func FileLineCount(fileName string) (int, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}

	maxReadSize := 1024 * 1024 // 1MB buffer size
	buffer := make([]byte, 0, maxReadSize)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(buffer, maxReadSize)

	count := 0
	for scanner.Scan() {
		count++
	}

	return count, nil
}

func FileLineCountWithLineSize(fileName string, lineSize int) (int, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}

	maxReadSize := lineSize
	buffer := make([]byte, 0, maxReadSize)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(buffer, maxReadSize)

	count := 0
	for scanner.Scan() {
		count++
	}

	return count, nil
}
