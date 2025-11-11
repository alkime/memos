package device

import (
	"github.com/gen2brain/malgo"
)

type AudioDeviceConfig struct {
	Format           malgo.FormatType
	CaptureChannels  int
	PlaybackChannels int
	SampleRate       int
}
