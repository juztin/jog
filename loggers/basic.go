// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package loggers contains various jog.Logger implementations
package loggers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"code.minty.io/config"
	"code.minty.io/jog"
)

type basic struct {
	client    *http.Client
	url, name string
}

func timeoutFn(seconds int) func(string, string) (net.Conn, error) {
	d := time.Duration(seconds)
	timeout := time.Duration(d * time.Second)
	return func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}
}

func cfg() (client *http.Client, name, url string) {
	tr := &http.Transport{}
	if b, ok := config.GroupBool("jog", "verifySSL"); ok {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: b}}
	}
	timeout, ok := config.GroupInt("job", "timeout")
	if !ok {
		timeout = 3
	}
	tr.Dial = timeoutFn(timeout)
	client = &http.Client{Transport: tr}
	name = config.RequiredGroupString("jog", "name")
	url = config.RequiredGroupString("jog", "url")
	return
}

// Log sends the data to an HTTP endpoint
func (l *basic) Log(m interface{}) (int, error) {
	// Marshal to JSON
	b, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}

	// Send it on it's way
	buf := bytes.NewBuffer(b)
	resp, err := l.client.Post(l.url, "application/json", buf)
	if err != nil {
		return 0, err
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, errors.New(fmt.Sprintf("received a `%d` from endpoint `%s` with data -> %s", resp.StatusCode, l.url, b))
	}
	return len(b), nil
}

// SetBasic sets the output of the log package so any logging is passed through a basic logger
func SetBasic() {
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
		return &basic{client, url + name, name}
	}
	return &basic{client, fmt.Sprintf("%s/%s", url, name), name}
}
