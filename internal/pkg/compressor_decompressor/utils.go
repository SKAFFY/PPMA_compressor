package compressor_decompressor

// buildCumFreqWithEscape строит cumFreq для символов 0..255 + escape (индекс 256)
func buildCumFreqWithEscape(freq map[byte]int, escFreq uint64) ([]uint64, uint64) {
	cum := make([]uint64, 258) // индексы 0..257, cum[257] = total
	cum[0] = 0
	var sumFreq uint64 = 0
	for i := 0; i < 256; i++ {
		f := uint64(freq[byte(i)])
		sumFreq += f
		cum[i+1] = sumFreq
	}
	total := sumFreq + escFreq
	cum[257] = total
	return cum, total
}
