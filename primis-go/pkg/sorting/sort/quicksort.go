package sorting

func partition(arr []int, low, high int) int {

	// NOTE  take the last number in the array
	pivot := arr[high]
	// NOTE  prepare value for next pivot, it is
	// NOTE  gonna be the last index before
	// NOTE  we found a bigger has occurred
	i := low

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			// NOTE we also switch digits in case
			// NOTE when j is a faster pointer
			// NOTE so it keeps skipping the values which are bigger than
			// Note last index, but rotate the one who are smaller,
			// NOTE because it remembers the position of last "bigger" value which is
			// NOTE last element
			arr[j], arr[i] = arr[i], arr[j]
			i++
		}
	}

	arr[i], arr[high] = arr[high], arr[i]
	return i
}
