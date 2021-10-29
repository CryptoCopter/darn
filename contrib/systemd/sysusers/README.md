# Declare system user & group

The included `systemd` service expects an existing user & group `darn:darn` to run. `systemd-sysusers` allows for declarative creation of users.

Place in `/usr/lib/sysusers.d/` if packaging, or in `/etc/sysusers.d/` if installing manually.
