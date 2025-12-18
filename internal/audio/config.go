package audio

import (
	"github.com/gen2brain/malgo"
)

type DeviceConfig struct {
	Format           malgo.FormatType
	CaptureChannels  int
	PlaybackChannels int
	SampleRate       int
}
