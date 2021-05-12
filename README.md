# Just so darn simple

`darn` is a simple HTTP file server on which users can upload their files without
login or authentication. All files have a maximum lifetime and are then deleted.


This software is a fork of [gosh](https://github.com/oxzi/gosh), which retains the encryption of stored data.

## Features

- Standalone HTTP web server, no additional server needed
- Store with both files and some metadata
- Only save uploader IP address for legal reasons, downloads are anonymous
- Content and all metadata are automatically deleted after expiration
- Content and relevant metadata (filename) are encrypted while at rest
  - This is not so much for the privacy of the shared data, but rather so that the administrator does not have to worry about what crap people are sharing...
- Configurable maximum lifetime and file size for uploads
- Replace or drop configured MIME types
- Simple upload via `curl`, `wget` or the like
- User manual available from the `/` page
- Uploads can specify their own shorter lifetime
- Burn after Reading: uploads can be deleted after the first download


## Installation

### From source

```bash
git clone https://github.com/CryptoCopter/darn.git
cd darn
go build ./cmd/darn
```

## Execution & Configuration

The snytax for `darn` is rather simple:

```bash
./darn /path/to/config.toml
```

### Configuration

A sample configuration might look like this:

```toml
[server]
# Listen address for the HTTP server
listen = ":8080"
# Contact E-Mail for abuses
contact = "abuse@example.org"
# MimeMap to substitute/drop MIMEs
mime-map = ""
log-level = "INFO"

[store]
# Path to the store directory
directory = "/path/to/my/store/dir"
# Size of chunks for large files
chunk-size = "1MiB"
# Maximum file size in bytes
max-filesize = "10MiB"
# Maximum lifetime for files
max-lifetime = "24h"
```

### Mime Map

The *MimeMap* file contains both substitutions or *drops* in each line and
could look as follows.

```
# Replace text/html with text/plain
text/html text/plain

# Drop PNGs, because reasons.
image/png DROP
```

## Posting

Files can be submitted via HTTP POST with common tools, e.g., with `curl`.

```bash
# Upload foo.png
curl -F 'file=@foo.png' http://example.org/

# Burn after reading:
curl -F 'file=@foo.png' -F 'burn=1' http://example.org/

# Set a custom expiry date, e.g., one day:
curl -F 'file=@foo.png' -F 'time=1d' http://example.org/

# Or all together:
curl -F 'file=@foo.png' -F 'time=1d' -F 'burn=1' http://example.org/
```

## Note on transport security

`darn` does not provide any TLS-functionality. If you want to use HTTPS (as I would strongly recommend), use a reverse-proxy.
