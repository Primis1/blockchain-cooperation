package sorting

func insertionSort(arr []int) {
	for i := 0; i < len(arr)-1; i++ {
		for j := i; j > 0 && arr[j-1] > arr[j]; j-- {
			arr[j], arr[j-1] = arr[j-1], arr[j]
		}
	}

}
