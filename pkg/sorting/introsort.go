package sorting

import (
	"blockchain/pkg/utils"
	"math"
)

func IntroSort(arr []int) {
	if arr == nil {
		utils.DisplayErr("\nArray in QUICKSORT is empty!\n")
	}

	maxDepth := int(2 * math.Log2(float64(len(arr))))
	introsort(arr, 0, len(arr)-1, maxDepth)
}

func introsort(arr []int, low, high, maxDepth int) {
	if high-low < 16 {
		insertionSort(arr)
		return
	}

	if maxDepth == 0 {
		heapSort(arr[low : high+1])
	}

	// our king - quicksort

	p := partition(arr, low, high)
	introsort(arr, low, p-1, maxDepth-1)
	introsort(arr, p+1, high, maxDepth-1)
}
