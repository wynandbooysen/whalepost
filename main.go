package main

// whalepost
// Copyright (C) 2018 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"context"
	"flag"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/faryon93/handlers"
	"github.com/faryon93/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  constants
// ---------------------------------------------------------------------------------------

const (
	HttpCloseTimeout = 5 * time.Second
	HttpListen       = ":8000"
)

var (
	Token      string
	Endpoint   string
	ApiVersion string
	LabelAllow string
	ConfFile   string

	Config *Conf
)

// ---------------------------------------------------------------------------------------
//  application entry
// ---------------------------------------------------------------------------------------

func main() {
	var colors bool
	var err error
	flag.BoolVar(&colors, "colors", false, "force color logging")
	flag.StringVar(&Token, "token", "", "token for authentication")
	flag.StringVar(&Endpoint, "endpoint", "/var/run/docker.sock", "docker endpoint")
	flag.StringVar(&ApiVersion, "api", "1.36", "docker api version")
	flag.StringVar(&LabelAllow, "label", "whalepost.allow", "label to allow updates")
	flag.StringVar(&ConfFile, "conf", "~/.docker/config.json", "path to docker config")
	flag.Parse()

	// make sure all config options are set properly
	if Token == "" || Endpoint == "" || LabelAllow == "" || ApiVersion == "" {
		flag.Usage()
		return
	}

	// setup logger
	formater := logrus.TextFormatter{ForceColors: colors}
	logrus.SetFormatter(&formater)
	logrus.SetOutput(os.Stdout)
	logrus.Infoln("starting", GetAppVersion())

	// load the config file
	Config, err = LoadConf(ConfFile)
	if err != nil {
		logrus.Warnln("config file not loaded:", err.Error())
	}

	// setup http routes
	router := mux.NewRouter()
	router.Path("/robots.txt").HandlerFunc(handlers.NoRobots)
	r := router.PathPrefix("/api/v1").Subrouter()
	r.Methods(http.MethodPut).Path("/service/{ServiceId}").
		Handler(handlers.ChainFunc(ServiceUpdate, handlers.Keyed(Token)))

	// start the webserver
	srv := &http.Server{Addr: HttpListen, Handler: router}
	go func() {
		logrus.Println("http server is listening on", HttpListen)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorln("http server failed to start:", err.Error())
			return
		}
	}()
	defer func() {
		ctx, _ := context.WithTimeout(context.Background(), HttpCloseTimeout)
		srv.Shutdown(ctx)
		logrus.Infoln("http server shutdown completed")
	}()

	// wait for stop signals
	util.WaitSignal(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infoln("received SIGINT / SIGTERM going to shutdown")
}
