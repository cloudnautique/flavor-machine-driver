package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancher/flavor-machine-driver/driver"
)

func main() {
	if os.Getenv("RANCHER_MACHINE_DRIVER_DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
	}
	plugin.RegisterDriver(new(rancher.Driver))
}
