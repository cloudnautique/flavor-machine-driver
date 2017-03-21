CaaS Machine Driver
========

A Docker Machine driver for adding hosts in CaaS.

## Building

`make`


## Usage

The driver is configured through two directories containing YAML configuration files.

1. Providers directory

Configures Docker Machine fields for providers. An example might be to set the SSH user or userdata fields for a provider. The primary use case for this is to set fields that will be common to all flavor types for a provider. The location of this directory is set by the `PROVIDERS_DIR` environment variable.

An individual provider config is just a map of Docker Machine fields. One for Digital Ocean might look like the following (with filename `digitalocean.yml`).

```yaml
digitalocean-ssh-user: rancher
```

2. Flavors directory

Configures Docker Machine fields for a particular flavor type. The flavor configurations end up being presented to the user via the `rancher-flavor` Docker Machine field. The primary use case for this is to configure the list of flavors that are available for a user to choose when adding hosts. The location of this directory is set by the `FLAVORS_DIR` environment variable.

A flavor for Digital Ocean might look like the following.

```yaml
provider: digitalocean
driver_options:
  digitalocean-image: ubuntu-16-04-x64
```

The `provider` key corresponds to the filename of a provider config, `digitalocean.yml` in this case. Everything under `driver_options` are Docker Machine fields.

If a field is present in the configuration from both the flavors directory and the providers directory then preference is given to the field from the flavor configuration.

## License
Copyright (c) 2014-2016 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
