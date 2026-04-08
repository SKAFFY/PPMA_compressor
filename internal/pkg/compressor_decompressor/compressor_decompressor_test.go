package compressor_decompressor

import (
	"bytes"
	"io"
	"testing"

	"PPMC_compressor/internal/pkg/arithmetic_encoder_decoder"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		maxOrder int
	}{
		{"empty", []byte{}, 2},
		{"single byte 'A'", []byte("A"), 2},
		{"two same bytes", []byte("AA"), 2},
		{"two different bytes", []byte("AB"), 2},
		{"short text", []byte(".pn 0"), 3},
		{"repeating pattern", []byte("abcabcabc"), 3},
		{"paper5 header", []byte(".pn 0\n.EQ"), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Буфер для сжатых данных
			var compressed bytes.Buffer

			// Кодировщик пишет в compressed
			enc := arithmetic_encoder_decoder.NewArithmeticEncoder(&compressed)

			// Создаём компрессор, который пишет заголовок в тот же буфер
			comp, err := NewCompressor(&compressed, enc, tt.maxOrder, uint64(len(tt.input)))
			if err != nil {
				t.Fatalf("NewCompressor error: %v", err)
			}

			// Сжимаем
			_, err = comp.Write(tt.input)
			if err != nil {
				t.Fatalf("Compressor.Write error: %v", err)
			}
			err = comp.Close()
			if err != nil {
				t.Fatalf("Compressor.Close error: %v", err)
			}

			// Декомпрессор читает из того же буфера
			dec, err := NewDecompressor(&compressed)
			if err != nil {
				t.Fatalf("NewDecompressor error: %v", err)
			}

			// Распаковываем
			decompressed := make([]byte, len(tt.input))
			_, err = io.ReadFull(dec, decompressed)
			if err != nil && err != io.EOF {
				t.Fatalf("Decompressor.Read error: %v", err)
			}

			// Сравниваем
			if !bytes.Equal(tt.input, decompressed) {
				t.Errorf("round-trip mismatch:\n original: %q\n got:      %q", tt.input, decompressed)
			}
		})
	}
}
