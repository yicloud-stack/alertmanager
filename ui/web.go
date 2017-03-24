// Copyright 2015 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ui

import (
	"bytes"
	"io"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/route"
)

func serveAsset(w http.ResponseWriter, req *http.Request, fp string) {
	info, err := AssetInfo(fp)
	if err != nil {
		log.Warn("Could not get file: ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	file, err := Asset(fp)
	if err != nil {
		if err != io.EOF {
			log.With("file", fp).Warn("Could not get file: ", err)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.ServeContent(w, req, info.Name(), info.ModTime(), bytes.NewReader(file))
}

// Register registers handlers to serve files for the web interface.
func Register(r *route.Router, reloadCh chan<- struct{}) {
	ihf := prometheus.InstrumentHandlerFunc

	r.Get("/metrics", prometheus.Handler().ServeHTTP)

	r.Get("/", ihf("index", func(w http.ResponseWriter, req *http.Request) {
		serveAsset(w, req, "ui/app/index.html")
	}))

	r.Get("/script.js", ihf("app", func(w http.ResponseWriter, req *http.Request) {
		serveAsset(w, req, "ui/app/script.js")
	}))

	r.Post("/-/reload", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Reloading configuration file..."))
		reloadCh <- struct{}{}
	})

	r.Get("/debug/*subpath", http.DefaultServeMux.ServeHTTP)
	r.Post("/debug/*subpath", http.DefaultServeMux.ServeHTTP)
}
