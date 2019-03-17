package converter

// TODO: Implement URLs?

// Converter is the generic interface for a audio and video converter
type Converter interface {
	Convert(inputVideoPath string, inputAudioPath string, outputVideoPath string) (err error)
}
