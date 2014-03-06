// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package loggers contains various jog.Logger implementations
package loggers

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"bitbucket.org/juztin/config"
	"bitbucket.org/juztin/jog"
)

type logger struct {
	client    *http.Client
	url, name string
}

// Log sends the data to an HTTP endpoint
func (l *logger) Log(p []byte) error {
	buf := bytes.NewBuffer(p)
	resp, err := l.client.Post(l.url, "application/json", buf)
	if err != nil {
		return err
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("received a `%d` from endpoint `%s` with data -> %s", resp.StatusCode, l.url, p))
	}
	return nil
}

func cfg() (client *http.Client, name, url string) {
	tr := &http.Transport{}
	if b, ok := config.GroupBool("jog", "verifySSL"); ok {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: b}}
	}

	client = &http.Client{Transport: tr}
	name = config.Required.GroupString("jog", "name")
	url = config.Required.GroupString("jog", "url")

	return
}

// SetBasicLogger sets the output of the log package so any logging is passed through a basic logger
func SetBasicLogger() {
	log.SetPrefix("")
	log.SetFlags(0)
	log.SetOutput(jog.NewWriter(New(cfg())))
}

// NewFromConfig returns a new basic jog.Logger using `jog` values from `config.json`
func NewFromConfig() jog.Logger {
	return New(cfg())
}

// New returns a new basic jog.Logger
func New(client *http.Client, name, url string) jog.Logger {
	if strings.HasSuffix(url, "/") {
		return &logger{client, name, url + name}
	}
	return &logger{client, name, fmt.Sprintf("%s/%s", url, name)}
}
