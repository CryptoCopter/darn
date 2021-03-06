package internal

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/akamensky/base58"

	log "github.com/sirupsen/logrus"
)

const indexTpl = `<!DOCTYPE html>
<html>
	<head>
		<style>
			* {
				font-family: monospace;
			}

			body {
				margin: 0 auto;
				padding: 1rem;
				width: 50%;
			}

			h1 {
				padding-top: 3rem;
			}

			h2 {
				padding-top: 2rem;
			}

			h3 {
				padding-top: 1rem;
			}

			pre {
				background-color: #eee;
				padding: 0.5rem;
			}

			form {
				padding: 0.5rem;
				position: relative;
				margin: auto;
				background-color: #eee;
			}

			#grid {
				display: grid;
				grid-gap: 1rem;
				grid-template-columns: 1fr 1fr;
				grid-template-rows: repeat(3, 3rem);
				margin-bottom: 1rem;
			}

			#grid > * {
				margin: auto 0;
			}

			#grid input[type="checkbox"] {
				margin-right: auto;
			}

			button {
				width: 100%;
			}
		</style>

		<title>darn!</title>
	</head>

	<body>
		<h1># Just so darn simple</h1>
		<p>
			Upload your files to this server and share them with your friends or, if
			non-existent, shady people from the Internet.
		</p>
		<p>
			Your file will expire after {{.Expires}} or earlier, if explicitly
			specified. Optionally, the file can be deleted directly after the first
			retrieval. In addition, the maximum file size is {{.Size}}.
		</p>
		<p>
			This is no place to share questionable or illegal data. Please use another
			service or stop it completely. Get some help.
		</p>
		<p>
			The source can be obtained from
			<a href="https://github.com/CryptoCopter/darn">https://github.com/CryptoCopter/darn</a>
		</p>

		<h2>## Posting</h2>

		<h3>### curl</h3>

		HTTP POST your file:

		<pre>$ curl -F 'file=@foo.png' {{.Proto}}://{{.Hostname}}/</pre>

		Burn after reading:

		<pre>$ curl -F 'file=@foo.png' -F 'burn=1' {{.Proto}}://{{.Hostname}}/</pre>

		Set a custom expiry date, e.g., one minute:

		<pre>$ curl -F 'file=@foo.png' -F 'time=1m' {{.Proto}}://{{.Hostname}}/</pre>

		Or all together:

		<pre>$ curl -F 'file=@foo.png' -F 'time=1m' -F 'burn=1' {{.Proto}}://{{.Hostname}}/</pre>

		<h3>### form</h3>

		<form
			action="{{.Proto}}://{{.Hostname}}/"
			method="POST"
			enctype="multipart/form-data">
			<div id="grid">
				<label for="file">Your file:</label>
				<input type="file" name="file" />
				<label for="burn">Burn after reading:</label>
				<input type="checkbox" name="burn" value="1" />
				<label for="time">Optionally, set a custom expiry date:</label>
				<input
					type="text"
					name="time"
					pattern="{{.DurationPattern}}"
					title="A duration string is sequence of decimal numbers, each with a unit suffix. Valid time units in order are 'y', 'mo', 'w', 'd', 'h', 'm', 's'"
				/>
			</div>
			<button>Upload</button>
		</form>

		<h2>## Retrieval & Deletion</h2>

		After a successful upload, you will receive a URL

		<pre>{{.Proto}}://{{.Hostname}}/TOKEN</pre>

		you can send a GET-request to retrieve the contents or a DELETE-request to have the server remove the content immediately.

		<pre>$ curl -X DELETE {{.Proto}}://{{.Hostname}}/TOKEN</pre>

		Why is there no HTML form to do that part? Because modifying the request URL of a from on the fly would require javascript.

		<h2>## Privacy</h2>

		This software stores the IP address for each upload. This information is
		stored as long as the file is available. In addition, the IP address of the
		user might be logged in case of an error. A normal download is logged without
		user information.

		<h2>## Abuse</h2>

		If, for whatever reason, you would like to have a file removed prematurely,
		please write an e-mail to
		<a href="mailto:{{.EMail}}">&lt;{{.EMail}}&gt;</a>. Please allow me a
		certain amount of time to react and work on your request.
	</body>
</html>
`

const (
	msgFileSizeExceeds = "Error: File size exceeds maximum."
	msgGenericError    = "Error: Something went wrong."
	msgIllegalMime     = "Error: MIME type is blacklisted."
	// msgIllegalMeme       = "Error: Content banned by EU Article 13"
	msgLifetimeExceeds   = "Error: Lifetime exceeds maximum."
	msgNotExists         = "Error: Does not exist."
	msgUnsupportedMethod = "Error: Method not supported."
)

// Server implements a http.Handler for up- and download.
type Server struct {
	store       *Store
	maxSize     int64
	maxLifetime time.Duration
	contactMail string
	mimeMap     MimeMap
	chunkSize   uint64
}

// NewServer creates a new Server with a given database directory, and
// configuration values. The Server must be started as a http.Handler.
func NewServer(storeDirectory string, maxSize int64, maxLifetime time.Duration,
	contactMail string, mimeMap MimeMap, chunkSize uint64) (s *Server, err error) {
	store, storeErr := NewStore(storeDirectory, true)
	if storeErr != nil {
		err = storeErr
		return
	}

	s = &Server{
		store:       store,
		maxSize:     maxSize,
		maxLifetime: maxLifetime,
		contactMail: contactMail,
		mimeMap:     mimeMap,
		chunkSize:   chunkSize,
	}
	return
}

// Close the Server and its components.
func (serv *Server) Close() error {
	return serv.store.Close()
}

func (serv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if reqPath := r.URL.Path; reqPath == "/" {
		serv.handleRoot(w, r)
	} else {
		serv.handleItem(w, r)
	}
}

func (serv *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		serv.handleIndex(w, r)

	case http.MethodPost:
		serv.handleUpload(w, r)

	default:
		log.WithField("method", r.Method).Debug("Called with unsupported method")

		http.Error(w, msgUnsupportedMethod, http.StatusMethodNotAllowed)
	}
}

func (serv *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index").Parse(indexTpl)
	if err != nil {
		log.WithError(err).Warn("Failed to parse template")

		http.Error(w, msgGenericError, http.StatusBadRequest)
		return
	}

	data := struct {
		Expires         string
		Size            string
		Proto           string
		Hostname        string
		EMail           string
		DurationPattern string
	}{
		Expires:         PrettyDuration(serv.maxLifetime),
		Size:            PrettyBytesize(serv.maxSize),
		Proto:           WebProtocol(r),
		Hostname:        r.Host,
		EMail:           serv.contactMail,
		DurationPattern: getHtmlDurationPattern(),
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := t.Execute(w, data); err != nil {
		log.WithError(err).Warn("Failed to execute template")
	}
}

func (serv *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	item, f, err := NewItem(r, serv.maxSize, serv.maxLifetime, serv.chunkSize)

	if err != nil {
		switch err {
		case ErrLifetimeToLong:
			log.Info("New Item with a too great lifetime was rejected")

			http.Error(w, msgLifetimeExceeds, http.StatusNotAcceptable)
			return
		case ErrFileToBig:
			log.Info("New Item with a too great file size was rejected")

			http.Error(w, msgFileSizeExceeds, http.StatusNotAcceptable)
			return
		default:
			log.WithError(err).Warn("Failed to create new Item")

			http.Error(w, msgGenericError, http.StatusBadRequest)
			return
		}
	} else if serv.mimeMap.MustDrop(item.ContentType) {
		log.WithField("MIME", item.ContentType).Info("Prevented upload of an illegal MIME")

		http.Error(w, msgIllegalMime, http.StatusBadRequest)
		return
	}

	itemId, secretKey, err := serv.store.Put(item, f)
	if err != nil {
		log.WithError(err).Warn("Failed to store Item")

		http.Error(w, msgGenericError, http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"ID":      itemId,
		"expires": item.Expires,
	}).Info("Uploaded new Item")

	// encode the itemid and the secretkey into the returned URL
	tokenBytes, _ := base58.Decode(itemId)
	tokenBytes = append(tokenBytes, secretKey[:]...)
	token := base58.Encode(tokenBytes)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s://%s/%s\n", WebProtocol(r), r.Host, token)
}

func (serv *Server) handleItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		serv.handleRequest(w, r)

	case http.MethodDelete:
		serv.handleDelete(w, r)

	default:
		log.WithField("method", r.Method).Debug("Called with unsupported method")

		http.Error(w, msgUnsupportedMethod, http.StatusMethodNotAllowed)
	}
}

func (serv *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	reqId, secretKey, err := parseURL(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Debug("Malformed token")

		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}

	item, err := serv.store.GetDecrypted(reqId, secretKey, true)

	if err == ErrNotFound {
		log.WithField("ID", reqId).Debug("Requested non-existing ID")

		http.Error(w, msgNotExists, http.StatusNotFound)
		return
	} else if err != nil {
		log.WithError(err).WithField("ID", reqId).Warn("Requesting errored")

		http.Error(w, msgGenericError, http.StatusBadRequest)
		return
	}

	if f, err := serv.store.GetFile(item, secretKey); err != nil {
		log.WithError(err).WithField("ID", item.ID).Warn("Reading file errored")

		http.Error(w, msgGenericError, http.StatusBadRequest)
		return
	} else {
		mimeType, mimeErr := serv.mimeMap.Substitute(item.ContentType)
		if mimeErr != nil {
			log.WithError(err).WithField("ID", item.ID).Warn("Substituting MIME errored")

			http.Error(w, msgGenericError, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("inline; filename=\"%s\"", item.Filename))
		w.WriteHeader(http.StatusOK)

		if _, err := io.Copy(w, f); err != nil {
			// This might happen if the peer resets the connection, e.g., if
			// curl tries to print a non text file to stdout.
			log.WithError(err).WithField("ID", item.ID).Warn("Writing file errored")
		}

		if err := f.Close(); err != nil {
			log.WithError(err).WithField("ID", item.ID).Warn("Closing file errored")
		}

		log.WithFields(log.Fields{
			"ID": item.ID,
		}).Info("Item was requested")
	}

	if item.BurnAfterReading {
		log.WithField("ID", item.ID).Info("Item will be burned")
		if err := serv.store.Delete(item); err != nil {
			log.WithError(err).WithField("ID", item.ID).Warn("Deletion errored")
		}
	}
}

func (serv *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	reqId, secretKey, err := parseURL(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Debug("Malformed token")

		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}

	item, err := serv.store.GetDecrypted(reqId, secretKey, true)

	if err == ErrNotFound {
		log.WithField("ID", reqId).Debug("Requested non-existing ID")

		http.Error(w, msgNotExists, http.StatusNotFound)
		return
	} else if err != nil {
		log.WithError(err).WithField("ID", reqId).Warn("Requesting errored")

		http.Error(w, msgGenericError, http.StatusBadRequest)
		return
	}

	err = serv.store.Delete(item)
	if err != nil {
		log.WithError(err).Error("Error while attempting to delete item.")
		http.Error(w, msgGenericError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Item deleted")
}

// parseURL takes the request url and parses the item-id and secret key (if encryption is turned on)
func parseURL(path string) (string, [KeySize]byte, error) {
	var reqId string
	var secretKey [KeySize]byte

	token := strings.TrimLeft(path, "/")

	tokenBytes, err := base58.Decode(token)
	if err != nil {
		return "", [KeySize]byte{}, err
	}

	if len(tokenBytes) != (IDSize + KeySize) {
		return "", [KeySize]byte{}, fmt.Errorf("token has wrong size, expected size %v, got size %v", IDSize+KeySize, len(tokenBytes))
	}

	// partition the token into the request ID (first 4 bytes) and the secret key (last 32 bytes)
	reqId = base58.Encode(tokenBytes[:4])
	key := tokenBytes[4:]
	copy(secretKey[:], key)

	return reqId, secretKey, nil
}

// WebProtocol returns "http" or "https", based on the X-Forwarded-Proto header.
func WebProtocol(r *http.Request) string {
	if xfwp := r.Header.Get("X-Forwarded-Proto"); xfwp != "" {
		return xfwp
	} else {
		return "http"
	}
}
