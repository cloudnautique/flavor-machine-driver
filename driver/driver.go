package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/packethost/docker-machine-driver-packet"
)

const (
	flagPrefix = "rancher-"
)

var (
	flavorsDirDefault   = "/machine/flavors"
	providersDirDefault = "/machine/providers"
)

func init() {
	home := os.Getenv("CATTLE_HOME")
	if home == "" {
		home = "/var/lib/cattle"
	}
	flavorsDirDefault = home + flavorsDirDefault
	providersDirDefault = home + providersDirDefault
}

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
	DriverOptions map[string]interface{} `yaml:"driver_options,omitempty"`
}

var _ drivers.Driver = &Driver{}

type Driver struct {
	*drivers.BaseDriver

	ProviderDriverOptions map[string]interface{}

	AvailableFlavors map[string]Flavor
	flavor           Flavor

	AmazonEC2Driver    *amazonec2.Driver
	DigitalOceanDriver *digitalocean.Driver
	PacketDriver       *packet.Driver
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

// Transforms a list of flags to add rancher- as a prefix
func addPrefixToFlags(flags []mcnflag.Flag) []mcnflag.Flag {
	var newFlags []mcnflag.Flag
	for _, flag := range flags {
		switch flag := flag.(type) {
		case mcnflag.BoolFlag:
			flag.Name = flagPrefix + flag.Name
			newFlags = append(newFlags, flag)
		case mcnflag.IntFlag:
			flag.Name = flagPrefix + flag.Name
			newFlags = append(newFlags, flag)
		case mcnflag.StringFlag:
			flag.Name = flagPrefix + flag.Name
			newFlags = append(newFlags, flag)
		case mcnflag.StringSliceFlag:
			flag.Name = flagPrefix + flag.Name
			newFlags = append(newFlags, flag)
		}
	}
	return newFlags
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

	// Docker Machine requires all driver flags to be prefixed with rancher-
	return addPrefixToFlags(flags)
}

func (d *Driver) setupInnerDriver() error {
	if d.AmazonEC2Driver == nil {
		d.AmazonEC2Driver = amazonec2.NewDriver(d.MachineName, d.StorePath)
	}
	if d.DigitalOceanDriver == nil {
		d.DigitalOceanDriver = digitalocean.NewDriver(d.MachineName, d.StorePath)
	}
	if d.PacketDriver == nil {
		d.PacketDriver = packet.NewDriver(d.MachineName, d.StorePath)
	}

	if d.flavor.Provider == "amazonec2" {
		d.Driver = d.AmazonEC2Driver
	} else if d.flavor.Provider == "digitalocean" {
		d.Driver = d.DigitalOceanDriver
	} else if d.flavor.Provider == "packet" {
		d.Driver = d.PacketDriver
	}

	return nil
}

func getenv(name, def string) string {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	return v
}

func (d *Driver) readProviderAndFlavorInfo(selectedFlavor string) error {
	flavorsDir := getenv("FLAVORS_DIR", flavorsDirDefault)
	log.Debugf("Reading flavor configs from directory %s", flavorsDir)
	providersDir := getenv("PROVIDERS_DIR", providersDirDefault)
	log.Debugf("Reading provider configs from directory %s", providersDir)

	files, err := ioutil.ReadDir(flavorsDir)
	if err != nil {
		return fmt.Errorf("Failed to read flavors directory %s: %v", flavorsDir, err)
	}

	flavorFound := false
	selectedFlavorFile := selectedFlavor + ".yaml"
	for _, file := range files {
		if file.Name() == selectedFlavorFile {
			bytes, err := ioutil.ReadFile(path.Join(flavorsDir, file.Name()))
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(bytes, &d.flavor); err != nil {
				return err
			}
			flavorFound = true
			break
		}
	}
	if !flavorFound {
		return fmt.Errorf("Invalid flavor %s, did not file %s", selectedFlavor, selectedFlavorFile)
	}

	files, err = ioutil.ReadDir(providersDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("Failed to read providers directory: %v", err)
	}

	providerFile := d.flavor.Provider + ".yaml"
	for _, file := range files {
		if file.Name() == providerFile {
			bytes, err := ioutil.ReadFile(path.Join(providersDir, file.Name()))
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(bytes, &d.ProviderDriverOptions); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// Strips off the rancher- prefix from flags
func stripPrefixFromFlags(flags rpcdriver.RPCFlags) rpcdriver.RPCFlags {
	for k, v := range flags.Values {
		s := fmt.Sprint(k)
		if strings.Contains(s, flagPrefix) {
			delete(flags.Values, k)
			flags.Values[strings.Replace(s, flagPrefix, "", -1)] = v
		}
	}
	return flags
}

// Merge the four sources of flag values, from lowest priority to highest
// Defaults from inner driver
// Values determined by the provider
// Values determined by flavor
// Values passed in via CLI (likely API keys)
func getDriverOpts(mcnflags []mcnflag.Flag, providerDriverOptions, flavorDriverOptions, cliDriverOptions map[string]interface{}) rpcdriver.RPCFlags {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}
	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}
	for k, v := range providerDriverOptions {
		driverOpts.Values[k] = v
	}
	for k, v := range flavorDriverOptions {
		driverOpts.Values[k] = v
	}
	for k, v := range cliDriverOptions {
		driverOpts.Values[k] = v
	}

	// Strip off the rancher- prefix since inner drivers won't recognize it
	return stripPrefixFromFlags(driverOpts)
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	if err := d.readProviderAndFlavorInfo(flags.String("rancher-flavor")); err != nil {
		return err
	}

	if err := d.setupInnerDriver(); err != nil {
		return err
	}

	// TODO: try to avoid this type assertion
	cliDriverOptions := flags.(*rpcdriver.RPCFlags)
	driverOptions := getDriverOpts(d.Driver.GetCreateFlags(), d.ProviderDriverOptions, cliDriverOptions.Values, d.flavor.DriverOptions)

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
