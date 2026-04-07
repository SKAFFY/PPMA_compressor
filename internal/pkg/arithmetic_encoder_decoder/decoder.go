package arithmetic_encoder_decoder

import (
	"io"
)

type ArithmeticDecoder struct {
	low, high uint64
	value     uint64
	in        io.Reader
	buf       byte
	bits      uint8
	err       error
}

func NewArithmeticDecoder(r io.Reader) (*ArithmeticDecoder, error) {
	d := &ArithmeticDecoder{
		low:  0,
		high: TopValue - 1,
		in:   r,
	}
	for i := 0; i < CodeValueBits; i++ {
		bit, err := d.readBit()
		if err != nil {
			return nil, err
		}
		d.value = (d.value << 1) | bit
	}
	return d, nil
}

func (d *ArithmeticDecoder) readBit() (uint64, error) {
	if d.bits == 0 {
		var b [1]byte
		_, err := d.in.Read(b[:])
		if err != nil {
			return 0, err
		}
		d.buf = b[0]
		d.bits = 8
	}
	d.bits--
	return uint64((d.buf >> d.bits) & 1), nil
}

func (d *ArithmeticDecoder) Decode(cumFreq []uint64, totalFreq uint64) (byte, error) {
	if d.err != nil {
		return 0, d.err
	}
	rng := d.high - d.low + 1
	scaled := ((d.value-d.low+1)*totalFreq - 1) / rng

	lo, hi := 0, len(cumFreq)-2
	for lo <= hi {
		mid := (lo + hi) / 2
		if cumFreq[mid] <= scaled && scaled < cumFreq[mid+1] {
			sym := byte(mid)
			d.high = d.low + (rng*cumFreq[sym+1])/totalFreq - 1
			d.low = d.low + (rng*cumFreq[sym])/totalFreq

			for {
				if d.high < Half {
					// бит не меняется
				} else if d.low >= Half {
					d.value -= Half
					d.low -= Half
					d.high -= Half
				} else if d.low >= FirstQtr && d.high < ThirdQtr {
					d.value -= FirstQtr
					d.low -= FirstQtr
					d.high -= FirstQtr
				} else {
					break
				}
				d.low <<= 1
				d.high = (d.high << 1) | 1
				bit, err := d.readBit()
				if err != nil {
					d.err = err
					return 0, err
				}
				d.value = (d.value << 1) | bit
			}
			return sym, nil
		} else if scaled < cumFreq[mid] {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return 0, io.ErrUnexpectedEOF
}
