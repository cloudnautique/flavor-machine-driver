package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancherlabs/caas-machine-driver/driver"
)

func main() {
	plugin.RegisterDriver(new(rancher.Driver))
}
