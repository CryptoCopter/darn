# gosh! Go Share [![Build Status](https://travis-ci.org/oxzi/gosh.svg?branch=master)](https://travis-ci.org/oxzi/gosh)

gosh is a simple HTTP file server on which users can upload their files without
login or authentication. All files have a maximum lifetime and are then deleted.


## Features

- Standalone HTTP web server, no additional server needed
- Store with both files and some metadata
- Only safe uploader's IP address for legal reasons, anonymous download
- File and all metadata are automatically deleted after expiration
- Configurable maximum lifetime and file size for uploads
- Replace or drop configured MIME types
- Simple upload via `curl`, `wget` or the like
- User manual available from the `/` page
- Uploads can specify their own shorter lifetime
- Burn after Reading: uploads can be deleted after the first download


## Installation

```bash
git clone https://github.com/oxzi/gosh.git
cd gosh

go build ./cmd/goshd
go build ./cmd/gosh-query
```


## Commands
### goshd

`goshd` is the web server, as described above.

```
Usage of ./goshd:
  -contact string
        Contact E-Mail for abuses
  -listen string
        Listen address for the HTTP server (default ":8080")
  -max-filesize string
        Maximum file size in bytes (default "10MiB")
  -max-lifetime string
        Maximum lifetime (default "24h")
  -mimemap string
        MimeMap to substitute/drop MIMEs
  -store string
        Path to the store
  -verbose
        Verbose logging
```

An example usage could look like this.

```bash
./goshd \
  -contact my@email.address \
  -max-filesize 64MiB \
  -max-lifetime 2w \
  -mimemap Mimemap \
  -store /path/to/my/store/dir
```

The *MimeMap* file contains both substitutions or *drops* in each line and
could look as follows.

```
# Replace text/html with text/plain
text/html text/plain

# Drop PNGs, because reasons.
image/png DROP
```


### gosh-query

The store can also be queried offline to get information or delete items. This
is `gosh-query`'s job.

```
Usage of ./gosh-query:
  -delete
        Delete selection
  -id string
        Query for an ID
  -ip-addr string
        Query for an IP address
  -store string
        Path to the store, env variable GOSHSTORE can also be used
  -verbose
        Verbose logging
```

```bash
# Find all uploads from the localhost:
./gosh-query -store /path/to/store -ip-addr ::1

# Show information for upload with ID abcdef
./gosh-query -store /path/to/store -id abcdef

# And delete this one
./gosh-query -store /path/to/store -delete -id abcdef
```

## Posting

Files can be submitted via HTTP POST with common tools, e.g., with `curl`.

```bash
# Upload foo.png
curl -F 'file=@foo.png' http://our-server.example/

# Burn after reading:
curl -F 'file=@foo.png' -F 'burn=1' http://our-server.example/

# Set a custom expiry date, e.g., one day:
curl -F 'file=@foo.png' -F 'time=1d' http://our-server.example/

# Or all together:
curl -F 'file=@foo.png' -F 'time=1d' -F 'burn=1' http://our-server.example/
```
