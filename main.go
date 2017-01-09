package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancher/machine-driver/driver"
)

func main() {
	plugin.RegisterDriver(new(rancher.Driver))
}
