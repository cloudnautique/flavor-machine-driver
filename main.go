package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancher/machine-driver/driver"
)

func main() {
	rancherDriver := new(rancher.Driver)
	rancherDriver.AvailableFlavors = map[string]rancher.Flavor{
		"amazonec2": rancher.Flavor{
			Provider: "amazonec2",
			DriverOptions: map[string]interface{}{
				"amazonec2-region": "us-west-2",
				"amazonec2-ami":    "ami-b7a114d7",
			},
		},
		"digitalocean": rancher.Flavor{
			Provider: "digitalocean",
		},
		"packet": rancher.Flavor{
			Provider: "packet",
			DriverOptions: map[string]interface{}{
				"packet-os": "rancher",
			},
		},
	}
	plugin.RegisterDriver(rancherDriver)
}
