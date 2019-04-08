package config

// ConverterType is the type of converter
type ConverterType int

const (
	// ConverterFFmpeg is the type reserved for a FFmpeg converter
	ConverterFFmpeg ConverterType = iota
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
