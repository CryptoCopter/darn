# Just so darn simple

`darn` is a simple HTTP file server on which users can upload their files without
login or authentication. All files have a maximum lifetime and are then deleted.


This software is a fork of [gosh](https://github.com/oxzi/gosh), which retains the encryption of stored data.

## Features

- Standalone HTTP web server, no additional server needed
- Store with both files and some metadata
- Only save uploader's IP address for legal reasons, anonymous download
- File and all metadata are automatically deleted after expiration
- Configurable maximum lifetime and file size for uploads
- Replace or drop configured MIME types
- Simple upload via `curl`, `wget` or the like
- User manual available from the `/` page
- Uploads can specify their own shorter lifetime
- Burn after Reading: uploads can be deleted after the first download


## Installation

```bash
git clone https://github.com/CryptoCopter/darn.git
cd darn

go build ./cmd/darn
```


## Commands
### darn

`darn` is the web server, as described above.

```
Usage of ./darn:
  -chunk-size string
    	Size of chunks for large files (default "1MiB")
  -contact string
    	Contact E-Mail for abuses
  -encrypt
    	Encrypt stored data
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
./darn \
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
