package compressor_decompressor

import (
	"PPMA_compressor/internal/pkg/arithmetic_encoder_decoder"
	"PPMA_compressor/internal/pkg/context_tree"
	"PPMA_compressor/internal/pkg/sliding_window"
	"encoding/binary"
	"fmt"
	"io"
)

type Decompressor struct {
	decoder       *arithmetic_encoder_decoder.ArithmeticDecoder
	contextTree   *context_tree.ContextTree
	maxOrder      int
	slidingWindow *sliding_window.SlidingWindow
	remaining     uint64
	originalSize  uint64
	contextBuf    []byte // переиспользуемый буфер для контекста
}

// NewDecompressor читает заголовок из r и создаёт декомпрессор с арифметическим декодером.
func NewDecompressor(r io.Reader) (*Decompressor, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	maxOrder := int(header[0])
	originalSize := binary.LittleEndian.Uint64(header[1:])

	decoder := arithmetic_encoder_decoder.NewArithmeticDecoder(r)

	return &Decompressor{
		decoder:       decoder,
		contextTree:   context_tree.NewContextTree(maxOrder),
		maxOrder:      maxOrder,
		slidingWindow: sliding_window.NewSlidingWindow(maxOrder),
		remaining:     originalSize,
		originalSize:  originalSize,
		contextBuf:    make([]byte, maxOrder), // буфер для контекста
	}, nil
}

// OriginalSize возвращает исходный размер данных.
func (d *Decompressor) OriginalSize() uint64 {
	return d.originalSize
}

// Read реализует io.Reader – декомпрессия данных.
func (d *Decompressor) Read(p []byte) (n int, err error) {
	for n < len(p) && d.remaining > 0 {
		sym, err := d.decodeNextSymbol()
		if err != nil {
			return n, fmt.Errorf("failed to decode symbol: %w", err)
		}
		p[n] = byte(sym)
		n++
		d.remaining--

		// обновление модели
		fullContext := d.slidingWindow.GetContext(d.maxOrder, d.contextBuf[:0])
		d.contextTree.Update(byte(sym), fullContext)
		d.slidingWindow.Push(byte(sym))
	}

	if d.remaining == 0 && n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// decodeNextSymbol декодирует один символ (0..255) из арифметического потока.
// Возвращает символ или ошибку.
func (d *Decompressor) decodeNextSymbol() (int, error) {
	order := d.maxOrder
	context := d.slidingWindow.GetContext(order, d.contextBuf[:0])

	for order >= 0 {
		node := d.contextTree.GetNode(context)

		if node != nil && node.Total > 0 {
			// Узел существует и имеет статистику
			escapeFreq := uint64(len(node.Freq)) // метод C
			cum, total := GetCumFreqWithEscape(node.Freq, escapeFreq)

			sym, err := d.decoder.Decode(cum, total)
			PutCumFreq(cum) // немедленно возвращаем в пул

			if err != nil {
				return 0, fmt.Errorf("decode error at order %d: %w", order, err)
			}

			if sym != Escape {
				return sym, nil
			}
			// Escape – переходим к меньшему порядку
		} else {
			cum := make([]uint64, 258)
			cum[257] = 1

			sym, err := d.decoder.Decode(cum, 1)
			if err != nil {
				return 0, fmt.Errorf("escape decode error at order %d: %w", order, err)
			}

			if sym != Escape {
				return 0, fmt.Errorf("expected Escape (256), got %d at order %d", sym, order)
			}
			// Escape – переходим к меньшему порядку
		}

		order--
		if order >= 0 {
			context = d.slidingWindow.GetContext(order, d.contextBuf[:0])
		}
	}

	// order == -1: равномерное распределение по 256 символам
	uniformCum, uniformTotal := GetUniformCumFreq()
	sym, err := d.decoder.Decode(uniformCum, uniformTotal)
	if err != nil {
		return 0, fmt.Errorf("decode error at uniform order: %w", err)
	}
	return sym, nil
}
