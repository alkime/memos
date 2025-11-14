package device

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alkime/memos/pkg/collections"
	"github.com/gen2brain/malgo"
)

type AudioDevice interface {
	// EnumerateDevices lists available audio devices.
	// It ignores any device configuration passed in.
	EnumerateDevices(ctx context.Context) ([]Info, error)

	// Capture initializes the underlying device and allocates a data packet
	// channel which, when Start() is called, will start receiving audio from
	// that device and writing packets of sampled bytes into the channel.
	Capture(ctx context.Context) (<-chan DataPacket, error)

	Start(ctx context.Context) error
	// Stop stops the audio device and optionally deallocates the underlying resources.
	// if the underlying device has already been deallocated this is a no-op.
	Stop(ctx context.Context, dealloc bool) error
}

type device struct {
	conf *AudioDeviceConfig

	mgCtx    *malgo.AllocatedContext
	mgDevice *malgo.Device
	dataC    chan DataPacket
}

func NewAudioDevice(conf *AudioDeviceConfig) AudioDevice {
	return &device{conf: conf}
}

func (d *device) EnumerateDevices(ctx context.Context) ([]Info, error) {
	// Initialize an empty context. AFAICT this is fine for just
	// enumrating the available devices.
	devCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize malgo context: %w", err)
	}
	defer uninitializeContext(devCtx)

	captureDevices, err := devCtx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to get capture devices: %w", err)
	}

	return collections.Apply(captureDevices, malgoDeviceInfoToDeviceInfo), nil
}

func (d *device) Capture(ctx context.Context) (<-chan DataPacket, error) {
	var err error
	d.dataC = make(chan DataPacket, 64)

	d.mgCtx, d.mgDevice, err = d.allocMGDevice(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to create malgo capture device: %w", err)
	}

	return d.dataC, nil
}

func (d *device) Start(ctx context.Context) error {
	if d.mgDevice == nil {
		return fmt.Errorf("device nil. have you allocated and Capture()ed it?")
	}

	if d.mgDevice.IsStarted() {
		// noop
		return nil
	}

	err := d.mgDevice.Start()
	if err != nil {
		return fmt.Errorf("failed to start malgo device: %w", err)
	}

	return nil
}

func (d *device) Stop(ctx context.Context, dealloc bool) error {
	if d.mgDevice == nil {
		// noop
		return nil
	}

	if err := d.mgDevice.Stop(); err != nil {
		return fmt.Errorf("failed to stop malgo device: %w", err)
	}

	if dealloc {
		d.deallocMGDevice()
		close(d.dataC)
	}

	return nil
}

func (d *device) allocMGDevice(devType malgo.DeviceType) (*malgo.AllocatedContext, *malgo.Device, error) {
	if d.dataC == nil {
		return nil, nil, fmt.Errorf("data channel is nil. unable to allocate device")
	}

	mgCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {
		slog.Info("malgo audio device log", "msg", msg)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize malgo context: %w", err)
	}

	var devCnf malgo.DeviceConfig
	var callBacks malgo.DeviceCallbacks

	switch devType { //nolint:exhaustive // Only Capture is supported; others handled by default
	case malgo.Capture:
		devCnf = malgo.DefaultDeviceConfig(malgo.Capture)
		devCnf.Capture.Format = d.conf.Format
		devCnf.Capture.Channels = uint32(d.conf.CaptureChannels)
		devCnf.SampleRate = uint32(d.conf.SampleRate)

		callBacks = malgo.DeviceCallbacks{
			Data: func(_, samples []byte, framecount uint32) {
				d.dataC <- DataPacket(samples)
			},
		}

	// todo: playback (& duplex???)
	default:
		return nil, nil, fmt.Errorf("unsupported device type: %v", devType)
	}

	mgDevice, err := malgo.InitDevice(mgCtx.Context, devCnf, callBacks)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize malgo device: %w", err)
	}

	return mgCtx, mgDevice, nil
}

func (d *device) deallocMGDevice() {
	if d.mgDevice == nil {
		return
	}

	d.mgDevice.Uninit()
	d.mgCtx.Free()
	d.mgDevice = nil
	d.mgCtx = nil
}

type Info struct {
	Name        string
	IsDefault   bool
	FormatCount int
	Formats     []string
}

func malgoDeviceInfoToDeviceInfo(mdi malgo.DeviceInfo) Info {
	formats := make([]string, len(mdi.Formats))
	for i, mf := range mdi.Formats {
		formats[i] = fmt.Sprintf("(SampleSizeBytes: %d, Channels: %d, SampleRate: %d)",
			malgo.SampleSizeInBytes(mf.Format),
			mf.Channels, mf.SampleRate)
	}
	return Info{
		Name:        mdi.Name(),
		IsDefault:   mdi.IsDefault != 0,
		FormatCount: int(mdi.FormatCount),
		Formats:     formats,
	}
}

type DataPacket []byte

func uninitializeContext(deviceCtx *malgo.AllocatedContext) {
	if deviceCtx == nil {
		return
	}

	if err := deviceCtx.Uninit(); err != nil {
		slog.Error("failed to uninitialize malgo context", "error", err)
	}
	deviceCtx.Free()
}
