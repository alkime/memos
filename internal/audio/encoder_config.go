package audio

import "errors"

const (
	// DefaultBufferThreshold is 4KB = 2048 mono samples = 128ms @ 16kHz.
	DefaultBufferThreshold = 4096
	// DefaultSampleRate is 16kHz, the native sample rate for Whisper.
	DefaultSampleRate = 16000
	// DefaultChannels is mono (1 channel).
	DefaultChannels = 1
)

// EncoderConfig configures the MP3 streaming encoder.
type EncoderConfig struct {
	// SampleRate is the audio sample rate in Hz (default: 16000 for Whisper).
	SampleRate int

	// Channels is the number of audio channels (default: 1 for mono).
	// Note: Internally converted to stereo for shine-mp3 encoder workaround.
	Channels int

	// BufferThreshold is the number of PCM bytes to accumulate before encoding.
	// Default: 4096 bytes (2048 samples, ~128ms @ 16kHz).
	BufferThreshold int
}

// Validate returns an error if the config is invalid.
func (c EncoderConfig) Validate() error {
	if c.SampleRate <= 0 {
		return errors.New("sample rate must be positive")
	}

	if c.Channels != 1 {
		return errors.New("only mono (1 channel) is supported")
	}

	if c.BufferThreshold <= 0 {
		return errors.New("buffer threshold must be positive")
	}

	return nil
}

// WithDefaults returns a config with default values applied to zero fields.
func (c EncoderConfig) WithDefaults() EncoderConfig {
	if c.SampleRate == 0 {
		c.SampleRate = DefaultSampleRate
	}

	if c.Channels == 0 {
		c.Channels = DefaultChannels
	}

	if c.BufferThreshold == 0 {
		c.BufferThreshold = DefaultBufferThreshold
	}

	return c
}
