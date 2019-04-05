package converter

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// ffmpeg holds the information relating to the FFmpeg executable
type ffmpeg struct {
	Path string
	log  logger.Logger
}

// FFmpeg sets up an FFmpeg converter
func FFmpeg(cfg *config.ConverterFFmpegConfig, loggerHandle logger.Logger) (converter Converter, err error) {
	loggerHandle.Debugf("FFmpeg(cfg): %#v", cfg)

	ffmpegPath, err := getExecutablePath(cfg.Path)
	if err != nil {
		return
	}

	converter = &ffmpeg{
		Path: ffmpegPath,
		log:  loggerHandle,
	}

	return
}

// ConvertFiles is the method to convert the files
func (f *ffmpeg) Convert(ctx context.Context, inputVideoPath string, inputAudioPath string, outputVideoPath string) (err error) {
	ffmpegArguments := []string{
		"-y",
		"-i", inputVideoPath,
		// ffmpeg audio arguments will go here
		"-c:v", "copy",
		"-strict", "experimental",
		"-f", "mp4",
		outputVideoPath,
	}

	if inputAudioPath != "" {
		ffmpegArguments = arrayInject(
			ffmpegArguments,
			[]string{
				"-i", inputAudioPath,
				"-c:a", "aac",
			},
			3,
		)
	}

	ffmpegCommand := exec.Command(
		f.Path,
		ffmpegArguments...,
	)

	output, err := ffmpegCommand.CombinedOutput()
	if err != nil {
		args := strings.Join(ffmpegCommand.Args, " ")
		os.Stderr.WriteString(string(output))
		f.log.Errorf("Error encountered while running command: %s", args)
		return
	}

	return
}

// Init is the initialization routine
func (f *ffmpeg) Init(ctx context.Context) (err error) {
	return
}

// arrayInject is a helper function written by Alirus on StackOverflow in my
// inquiry to find a way to inject one array into another _elegantly_:
// https://stackoverflow.com/a/53647212/776896
// Proof on Go Playground: https://play.golang.org/p/x4ljcOU71Z6
func arrayInject(haystack, pile []string, at int) (result []string) {
	result = make([]string, len(haystack[:at]))
	copy(result, haystack[:at])
	result = append(result, pile...)
	result = append(result, haystack[at:]...)

	return result
}

// getExecutablePath will return the path (relative or full) to Ffmpeg
func getExecutablePath(executable string) (ffmpegPath string, err error) {
	ffmpegPath, err = exec.LookPath(executable)
	return
}
