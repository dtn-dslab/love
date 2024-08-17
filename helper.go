package love

import (
	device "github.com/dtn-dslab/love/pkg/device"
)

type LoVeHelperConfig struct {
	MaxGoRoutines int
}

type LoVeHelper struct {
	waitGroup chan struct{}
}

func NewLoVeHelper(config LoVeHelperConfig) *LoVeHelper {
	return &LoVeHelper{
		waitGroup: make(chan struct{}, config.MaxGoRoutines),
	}
}

func (h *LoVeHelper) startRoutine() {
	<-h.waitGroup
}

func (h *LoVeHelper) endRoutine() {
	h.waitGroup <- struct{}{}
}

func (h *LoVeHelper) AddVEth(veth *VEth) error {
	h.startRoutine()
	defer h.endRoutine()

	return veth.AddVEth()
}

func (h *LoVeHelper) DeleteVEth(veth *VEth) error {
	h.startRoutine()
	defer h.endRoutine()

	return veth.DeleteVEth()
}

func (h *LoVeHelper) OpenVEth(veth *VEth) error {
	h.startRoutine()
	defer h.endRoutine()

	return veth.OpenVEth()
}

func (h *LoVeHelper) OpenDevice(device *device.Device) error {
	h.startRoutine()
	defer h.endRoutine()

	return device.OpenDevice()
}

func (h *LoVeHelper) DeleteDevice(device *device.Device) error {
	h.startRoutine()
	defer h.endRoutine()

	return device.DeleteDevice()
}

func (h *LoVeHelper) SetDevice(device *device.Device) error {
	h.startRoutine()
	defer h.endRoutine()

	return device.SetDevice()
}

func (h *LoVeHelper) SetDeviceProperties(device *device.Device) error {
	h.startRoutine()
	defer h.endRoutine()

	return device.SetDeviceProperties()
}
