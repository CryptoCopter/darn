# `systemd` service unit

Use this to run `darn` as a systemd-service.

Place this file in either `/usr/lib/systemd/system/` if you are building a package, or in `/etc/systemd/system/` if you are installing manually.

If installing manually, you might need to run `systemctl daemon-reload` before starting the service.

The service expects the following things:

- The `dtnd` binary installed to `/usr/bin/`
- An existing user & group `darn:darn`
- An existing working directory in `/srv/darn`
- Configuration in `/etc/darn.toml`
