package sliding_window

// SlidingWindow хранит последние N байт (maxOrder) в циклическом буфере.
// Позволяет получать контекст заданной длины без аллокаций (переиспользует переданный срез).
type SlidingWindow struct {
	buf  []byte
	pos  int
	size int // сколько реально символов записано (не больше len(buf))
}

// NewSlidingWindow создаёт окно с максимальным размером maxOrder.
// Если maxOrder == 0, окно всегда будет пустым.
func NewSlidingWindow(maxOrder int) *SlidingWindow {
	if maxOrder < 0 {
		maxOrder = 0
	}
	return &SlidingWindow{
		buf:  make([]byte, maxOrder),
		pos:  0,
		size: 0,
	}
}

// Push добавляет байт в окно, вытесняя самый старый при необходимости.
func (w *SlidingWindow) Push(b byte) {
	if len(w.buf) == 0 {
		// окно нулевого размера – ничего не храним
		return
	}
	w.buf[w.pos] = b
	w.pos = (w.pos + 1) % len(w.buf)
	if w.size < len(w.buf) {
		w.size++
	}
}

// GetContext возвращает контекст длины order (не более w.size) в переданный срез dst.
// Возвращает срез (возможно, dst[:order]), который действителен до следующего вызова GetContext или Push.
// Если order == 0, возвращает пустой срез.
// Если order > w.size, возвращает контекст длины w.size.
func (w *SlidingWindow) GetContext(order int, dst []byte) []byte {
	if len(w.buf) == 0 || order == 0 {
		return dst[:0]
	}
	if order > w.size {
		order = w.size
	}
	// предполагаем, что cap(dst) >= order
	dst = dst[:order]
	for i := 0; i < order; i++ {
		idx := w.pos - order + i
		if idx < 0 {
			idx += len(w.buf)
		}
		dst[i] = w.buf[idx]
	}
	return dst
}

// GetContextLegacy – старый метод, возвращающий новый срез (аллокация).
// Сохранён для обратной совместимости с тестами.
// В новом коде используйте GetContext(order, dst).
func (w *SlidingWindow) GetContextLegacy(order int) []byte {
	if len(w.buf) == 0 {
		return []byte{}
	}
	if order > w.size {
		order = w.size
	}
	if order == 0 {
		return []byte{}
	}
	context := make([]byte, order)
	for i := 0; i < order; i++ {
		idx := w.pos - order + i
		if idx < 0 {
			idx += len(w.buf)
		}
		context[i] = w.buf[idx]
	}
	return context
}
