package rancher

import (
	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/packethost/docker-machine-driver-packet"
)

var _ drivers.Driver = &Driver{}

type Driver struct {
	*drivers.BaseDriver
	AmazonEC2Driver    *amazonec2.Driver
	DigitalOceanDriver *digitalocean.Driver
	PacketDriver       packet.Driver
	Driver             drivers.Driver
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) setupInnerDriver() {
	if d.AmazonEC2Driver == nil {
		d.AmazonEC2Driver = amazonec2.NewDriver(d.MachineName, d.StorePath)
	}

	if d.DigitalOceanDriver == nil {
		d.DigitalOceanDriver = digitalocean.NewDriver(d.MachineName, d.StorePath)
	}

	d.PacketDriver.MachineName = d.MachineName
	d.PacketDriver.StorePath = d.StorePath

	d.Driver = d.DigitalOceanDriver
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	d.setupInnerDriver()
	var flags []mcnflag.Flag
	flags = append(flags, d.AmazonEC2Driver.GetCreateFlags()...)
	flags = append(flags, d.DigitalOceanDriver.GetCreateFlags()...)
	flags = append(flags, d.PacketDriver.GetCreateFlags()...)
	return flags
}

func (d *Driver) DriverName() string {
	return "rancher"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.setupInnerDriver()
	return d.Driver.SetConfigFromFlags(flags)
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.Driver.GetSSHHostname()
}

func (d *Driver) PreCreateCheck() error {
	return d.Driver.PreCreateCheck()
}

func (d *Driver) Create() error {
	return d.Driver.Create()
}

func (d *Driver) GetURL() (string, error) {
	return d.Driver.GetURL()
}

func (d *Driver) GetIP() (string, error) {
	return d.Driver.GetIP()
}

func (d *Driver) GetState() (state.State, error) {
	return d.Driver.GetState()
}

func (d *Driver) Start() error {
	return d.Driver.Start()
}

func (d *Driver) Stop() error {
	return d.Driver.Stop()
}

func (d *Driver) Remove() error {
	return d.Driver.Remove()
}

func (d *Driver) Restart() error {
	return d.Driver.Restart()
}

func (d *Driver) Kill() error {
	return d.Driver.Kill()
}
