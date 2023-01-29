package store

// Factory defines the factory storage interface
type Factory interface {
	Devices() DeviceStore
}
