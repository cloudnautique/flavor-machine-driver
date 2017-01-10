package rancher

import (
	"fmt"
	"strings"

	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/packethost/docker-machine-driver-packet"
)

const (
	tag             = "Rancher Cloud"
	vpcCidnBlock    = "10.0.0.0/16"
	subnetCidnBlock = "10.0.0.0/24"
)

var (
	apiKeyFlagNames = []string{
		// Amazon
		"access-key",
		"secret-key",

		// Digital Ocean
		"access-token",

		// Packet
		"api-key",
		"project-id",
	}
)

type Flavor struct {
	Provider      string
	DriverOptions map[string]interface{}
}

var _ drivers.Driver = &Driver{}

type Driver struct {
	*drivers.BaseDriver

	AvailableFlavors map[string]Flavor
	flavor           Flavor

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

func (d *Driver) DriverName() string {
	return "rancher"
}

func (d *Driver) setupInnerDriver() error {
	fmt.Println("setupInnerDriver")
	if d.AmazonEC2Driver == nil {
		d.AmazonEC2Driver = amazonec2.NewDriver(d.MachineName, d.StorePath)
	}

	if d.DigitalOceanDriver == nil {
		d.DigitalOceanDriver = digitalocean.NewDriver(d.MachineName, d.StorePath)
	}

	d.PacketDriver.MachineName = d.MachineName
	d.PacketDriver.StorePath = d.StorePath

	if d.flavor.Provider == "amazonec2" {
		d.Driver = d.AmazonEC2Driver
	} else if d.flavor.Provider == "digitalocean" {
		d.Driver = d.DigitalOceanDriver
	} else if d.flavor.Provider == "packet" {
		d.Driver = &d.PacketDriver
	}

	return nil
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	// Rancher specific flags
	flags := []mcnflag.Flag{
		mcnflag.StringFlag{
			Name: "flavor",
		},
	}

	// Borrow API key flags from inner drivers
	var innerFlags []mcnflag.Flag
	innerFlags = append(innerFlags, d.AmazonEC2Driver.GetCreateFlags()...)
	innerFlags = append(innerFlags, d.DigitalOceanDriver.GetCreateFlags()...)
	innerFlags = append(innerFlags, d.PacketDriver.GetCreateFlags()...)
	for _, innerFlag := range innerFlags {
		for _, name := range apiKeyFlagNames {
			if strings.Contains(innerFlag.String(), name) {
				flags = append(flags, innerFlag)
			}
		}
	}

	return flags
}

// Merge the three sources of flag values, from lowest priority to highest
// Defaults from inner driver
// Values determined by flavor
// Values passed in via CLI (likely API keys)
func getDriverOpts(mcnflags []mcnflag.Flag, flavorDriverOptions, cliDriverOptions map[string]interface{}) rpcdriver.RPCFlags {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}
	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}
	for k, v := range flavorDriverOptions {
		driverOpts.Values[k] = v
	}
	for k, v := range cliDriverOptions {
		driverOpts.Values[k] = v
	}
	return driverOpts
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	var ok bool
	d.flavor, ok = d.AvailableFlavors[flags.String("flavor")]
	if !ok {
		return fmt.Errorf("Unrecognized flavor %s", flags.String("flavor"))
	}

	if err := d.setupInnerDriver(); err != nil {
		return err
	}

	// TODO: try to avoid this type assertion
	cliDriverOptions := flags.(*rpcdriver.RPCFlags)
	driverOptions := getDriverOpts(d.Driver.GetCreateFlags(), cliDriverOptions.Values, d.flavor.DriverOptions)

	if err := d.Driver.SetConfigFromFlags(driverOptions); err != nil {
		return err
	}

	if d.flavor.Provider == "amazonec2" {
		if err := d.setupAmazon(); err != nil {
			return err
		}
	}

	return nil
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
