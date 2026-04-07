package arithmetic_encoder_decoder

const (
	CodeValueBits = 32
	TopValue      = uint64(1) << CodeValueBits
	FirstQtr      = TopValue / 4
	Half          = TopValue / 2
	ThirdQtr      = 3 * TopValue / 4
)
