package utils

import (
	"errors"
	"strconv"
	"strings"
)

var (
	errInvalidIndex = errors.New("Index is not valid")
)

// parseIndexList converts a list of indices to their integer values
func ParseIndexList(indices []string) ([]int, error) {
	var listIndex []int
	for _, strIndex := range indices {
		if !strings.Contains(strIndex, "-") {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, errInvalidIndex
			}

			listIndex = append(listIndex, index)
			continue
		}

		parts := strings.Split(strIndex, "-")
		if len(parts) != 2 {
			return nil, errInvalidIndex
		}

		minIndex, errMin := strconv.Atoi(parts[0])
		maxIndex, errMax := strconv.Atoi(parts[1])
		if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
			return nil, errInvalidIndex
		}

		for i := minIndex; i <= maxIndex; i++ {
			listIndex = append(listIndex, i)
		}
	}
	return listIndex, nil
}
