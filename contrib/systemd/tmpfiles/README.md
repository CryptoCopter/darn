# Create working directory

The included `systemd` service expects an existing in `/srv/darn`. `systemd-tmpfiles` allows for the programmatic creation of this directory.

NOTE: Contrary to what one might expect based on the name, this is not only useful for temporary files but also permanent data.

Place in `/usr/lib/tmpfiles.d` if packaging, or in `/etc/tmpfiles.d` if installing manually.