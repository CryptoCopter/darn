package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/CryptoCopter/darn/internal"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Listen   string `toml:"listen"`
	Contact  string `toml:"contact"`
	MimeMap  string `toml:"mime-map"`
	LogLevel string `toml:"log-level"`
}

type StoreConfig struct {
	Directory   string `toml:"directory"`
	ChunkSize   string `toml:"chunk-size"`
	MaxFilesize string `toml:"max-filesize"`
	MaxLifetime string `toml:"max-lifetime"`
}

type daemonConfig struct {
	Server ServerConfig
	Store  StoreConfig
}

func readConfig(path string) (daemonConfig, error) {
	config := daemonConfig{}

	content, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(content, &config)
	return config, err
}

func parseConfig(config daemonConfig) (*internal.Server, error) {
	maxLifetime, err := internal.ParseDuration(config.Store.MaxLifetime)
	if err != nil {
		return nil, err
	}

	maxFilesize, err := internal.ParseBytesize(config.Store.MaxFilesize)
	if err != nil {
		return nil, err
	}

	var chunkSize uint64
	cs, err := internal.ParseBytesize(config.Store.ChunkSize)
	if err != nil {
		return nil, err
	} else {
		chunkSize = uint64(cs)
	}

	var mimeMap internal.MimeMap
	if config.Server.MimeMap == "" {
		mimeMap = make(internal.MimeMap)
	} else {
		if f, err := os.Open(config.Server.MimeMap); err != nil {
			log.WithError(err).Fatal("Failed to open MimeMap")
		} else if mm, err := internal.NewMimeMap(f); err != nil {
			log.WithError(err).Fatal("Failed to parse MimeMap")
		} else {
			err = f.Close()
			if err != nil {
				log.WithError(err).Fatal("Error closing file.")
			}
			mimeMap = mm
		}
	}

	server, err := internal.NewServer(
		config.Store.Directory, maxFilesize, maxLifetime, config.Server.Contact, mimeMap, chunkSize)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func webserver(server *internal.Server, listenAddr string) {
	webServer := &http.Server{
		Addr:    listenAddr,
		Handler: server,
	}

	go func() {
		log.WithField("listen", listenAddr).Info("Starting web server")

		if err := webServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithError(err).Fatal("Web server errored")
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	<-stopChan
	log.Info("Closing web server")

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	if err := webServer.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Failed to shutdown web server")
	}
	ctxCancel()
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s configuration.toml", os.Args[0])
	}

	config, err := readConfig(os.Args[1])
	if err != nil {
		log.WithError(err).Fatal("Error reading config")
	}

	logLevel, err := log.ParseLevel(config.Server.LogLevel)
	if err != nil {
		log.WithError(err).Fatal("Error setting log level")
	}
	log.SetLevel(logLevel)

	server, err := parseConfig(config)
	if err != nil {
		log.WithError(err).Fatal("Error parsing config")
	}

	webserver(server, config.Server.Listen)

	if err := server.Close(); err != nil {
		log.WithError(err).Fatal("Closing errored")
	}
}
