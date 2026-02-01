package huffman

// Change the population counts in a way that the consequent
// Huffman tree compression, especially its RLE-part, give smaller output.
func OptimizeHuffmanForRle(length int, good_for_rle []uint8, counts []uint32) {
	// 1) Let's make the Huffman code more compatible with rle encoding.
	for ; length >= 0; length-- {
		if length == 0 {
			return // All zeros.
		}
		if counts[length-1] != 0 {
			// Now counts[0..length - 1] does not have trailing zeros.
			break
		}
	}
	// 2) Let's mark all population counts that already can be encoded
	// with an rle code.
	{
		// Let's not spoil any of the existing good rle codes.
		// Mark any seq of 0's that is longer as 5 as a good_for_rle.
		// Mark any seq of non-0's that is longer as 7 as a good_for_rle.
		symbol := counts[0]
		stride := 0
		for i := 0; i < length+1; i++ {
			if i == length || counts[i] != symbol {
				if (symbol == 0 && stride >= 5) || (symbol != 0 && stride >= 7) {
					for k := 0; k < stride; k++ {
						good_for_rle[i-k-1] = 1
					}
				}
				stride = 1
				if i != length {
					symbol = counts[i]
				}
			} else {
				stride++
			}
		}
	}
	// 3) Let's replace those population counts that lead to more rle codes.
	{
		var stride, sum uint32 = 0, 0
		limit := counts[0]

		for i := 0; i < length+1; i++ {
			if i == length || good_for_rle[i] != 0 || (i != 0 && good_for_rle[i-1] != 0) ||
				!ValuesShouldBeCollapsedToStrideAverage(int(counts[i]), int(limit)) {
				if stride >= 4 || (stride >= 3 && sum == 0) {
					// The stride must end, collapse what we have, if we have enough (4).
					count := (sum + stride/2) / stride
					if count < 1 {
						count = 1
					}
					if sum == 0 {
						// Don't make an all zeros stride to be upgraded to ones.
						count = 0
					}
					for k := uint32(0); k < stride; k++ {
						// We don't want to change value at counts[i], // that is already belonging to the next stride. Thus - 1.
						counts[i-int(k)-1] = count
					}
				}
				stride = 0
				sum = 0
				if i < length-3 {
					// All interesting strides have a count of at least 4, // at least when non-zeros.
					limit =
						(counts[i] + counts[i+1] + counts[i+2] + counts[i+3] + 2) /
							4
				} else if i < length {
					limit = counts[i]
				} else {
					limit = 0
				}
			}
			stride++
			if i != length {
				sum += counts[i]
				if stride >= 4 {
					limit = (sum + stride/2) / stride
				}
			}
		}
	}
}
