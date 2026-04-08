package compressor_decompressor

import (
	"sync"
)

// Пул для слайсов cumFreq размером 258 (256 символов + escape)
var cumFreqPool = sync.Pool{
	New: func() interface{} {
		return make([]uint64, 258)
	},
}

// GetCumFreqWithEscape возвращает слайс из пула, заполненный накопительными частотами.
// После использования слайс необходимо вернуть вызовом PutCumFreq.
func GetCumFreqWithEscape(freq map[byte]int, escFreq uint64) (cum []uint64, totalFreq uint64) {
	cum = cumFreqPool.Get().([]uint64)
	var sum uint64 = 0
	for i := 0; i < 256; i++ {
		f := uint64(freq[byte(i)])
		sum += f
		cum[i+1] = sum
	}
	totalFreq = sum + escFreq
	cum[257] = totalFreq
	return cum, totalFreq
}

// PutCumFreq возвращает слайс в пул.
func PutCumFreq(cum []uint64) {
	if cap(cum) >= 258 {
		cumFreqPool.Put(cum)
	}
}

// Глобальный слайс для равномерного распределения (order = -1)
var (
	uniformCumFreq   []uint64
	uniformTotalFreq uint64
	uniformOnce      sync.Once
)

// GetUniformCumFreq возвращает константный слайс для равномерного распределения.
func GetUniformCumFreq() (cum []uint64, totalFreq uint64) {
	uniformOnce.Do(func() {
		uniformCumFreq = make([]uint64, 257)
		for i := 0; i < 256; i++ {
			uniformCumFreq[i+1] = uint64(i + 1)
		}
		uniformTotalFreq = 256
	})
	return uniformCumFreq, uniformTotalFreq
}
