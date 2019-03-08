package config

type ConverterType int

const (
	ConverterFmpeg ConverterType = iota
)

// ConverterConfig holds all the different implmentations for video conversion
type ConverterConfig struct {
	FFmpeg ConverterFFmpegConfig
	Type   ConverterType
}

// String will return the string representation of the iota
func (c ConverterType) String() string {
	return [...]string{"ffmpeg"}[c]
}
