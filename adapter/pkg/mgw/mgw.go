/*
 *  Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

// Package mgw contains the implementation to start the adapter
package mgw

import (
	"errors"
	"log"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xdsv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	apiservice "github.com/envoyproxy/go-control-plane/wso2/discovery/service/api"
	configservice "github.com/envoyproxy/go-control-plane/wso2/discovery/service/config"
	"github.com/wso2/micro-gw/pkg/api/restserver"

	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	"archive/zip"
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/wso2/micro-gw/config"
	logger "github.com/wso2/micro-gw/loggers"
	xds "github.com/wso2/micro-gw/pkg/xds"
	"google.golang.org/grpc"
)

var (
	debug       bool
	onlyLogging bool

	localhost = "0.0.0.0"

	port        uint
	gatewayPort uint
	alsPort     uint

	mode string
)

const (
	ads = "ads"
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.BoolVar(&onlyLogging, "onlyLogging", false, "Only demo AccessLogging Service")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.UintVar(&gatewayPort, "gateway", 18001, "Management server port for HTTP gateway")
	flag.UintVar(&alsPort, "als", 18090, "Accesslog server port")
	flag.StringVar(&mode, "ads", ads, "Management server type (ads, xds, rest)")
}

const grpcMaxConcurrentStreams = 1000000

func runManagementServer(server xdsv3.Server, enforcerServer xdsv3.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.LoggerMgw.Fatal("failed to listen: ", err)
	}

	// register services
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	configservice.RegisterConfigDiscoveryServiceServer(grpcServer, enforcerServer)
	apiservice.RegisterApiDiscoveryServiceServer(grpcServer, enforcerServer)

	logger.LoggerMgw.Info("port: ", port, " management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.LoggerMgw.Error(err)
		}
	}()
}

// Run starts the XDS server and Rest API server.
func Run(conf *config.Config) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	// TODO: (VirajSalaka) Support the REST API Configuration via flags only if it is a valid requirement
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// log config watcher
	// TODO: (VirajSalaka) Implement a rest endpoint to apply configurations
	watcherLogConf, _ := fsnotify.NewWatcher()
	errC := watcherLogConf.Add("conf/log_config.toml")

	if errC != nil {
		logger.LoggerMgw.Fatal("Error reading the log configs. ", errC)
	}

	logger.LoggerMgw.Info("Starting adapter ....")
	cache := xds.GetXdsCache()
	enforcerCache := xds.GetEnforcerCache()
	srv := xdsv3.NewServer(ctx, cache, nil)
	enforcerXdsSrv := xdsv3.NewServer(ctx, enforcerCache, nil)

	runManagementServer(srv, enforcerXdsSrv, port)

	// Set enforcer startup configs
	xds.UpdateEnforcerConfig(conf)

	barry, ferr := fetchAPIs()
	if ferr != nil {
		logger.LoggerMgw.Info("Error fetch APIs", ferr)
	} else {
		err := applyAPIProject(barry)
		if err != nil {
			logger.LoggerMgw.Info("Error occured while starting:!!!! ", err)
		}
	}
	logger.LoggerMgw.Info("DDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")
	go restserver.StartRestServer(conf)
OUTER:
	for {
		select {
		case l := <-watcherLogConf.Events:
			switch l.Op.String() {
			case "WRITE":
				logger.LoggerMgw.Info("Loading updated log config file...")
				config.ClearLogConfigInstance()
				logger.UpdateLoggers()
			}
		case s := <-sig:
			switch s {
			case os.Interrupt:
				logger.LoggerMgw.Info("Shutting down...")
				break OUTER
			}
		}
	}
	logger.LoggerMgw.Info("Bye!")
}

// fetchAPIs pulls the API artifact calling to the API manager
// API Manager returns a .zip file as a response and this function
// returns a byte array of that ZIP file.
func fetchAPIs() ([]byte, error) {
	// URL has to be taken from a config, config toml already has this
	url := "https://172.17.0.1:9443/internal/data/v1/runtime-artifacts"
	logger.LoggerMgw.Info("Starting URL ....")

	insecure := false
	// Create the request
	//handle TLS
	tr := &http.Transport{}
	if !insecure {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	// Configuring the http client
	client := &http.Client{
		Transport: tr,
	}

	// create a HTTP request
	req, err := http.NewRequest("GET", url, nil)
	// Adding query parameters
	q := req.URL.Query()
	// todo: parameterized the queries
	q.Add("gatewayLabel", "Production and Sandbox")
	q.Add("type", "Envoy")
	req.URL.RawQuery = q.Encode()
	logger.LoggerMgw.Info("Starting query ....")

	// Setting authorization header
	req.Header.Set("Authorization", "Basic YWRtaW46YWRtaW4")
	// Make the request
	resp, err := client.Do(req)
	// In the event of a connection error, the error would not be nil, then return the error
	// If the error is not null, proceed
	if err != nil {
		logger.LoggerMgw.Error("Error occurred while retrieving APIs from API manager:", err)
		return nil, err
	}
	logger.LoggerMgw.Info(resp.StatusCode)

	// For sucessful API retrieval, return the apis.zip in a form of byte array
	respbytes, err := ioutil.ReadAll(resp.Body)

	// If the reading response gives an error
	if err != nil {
		fmt.Println("Error occurred while reading the response.", err)
		return nil, err
	}
	// For successful response, return the byte array and nil as error
	if resp.StatusCode == http.StatusOK {
		return respbytes, nil
	}
	// If the response is not successful, create a new error with the response and log it and return
	// Ex: for 401 scenarios, 403 scenarios.
	logger.LoggerMgw.Info("Failure response code:", resp.StatusCode)
	logger.LoggerMgw.Info("Failure response:", string(respbytes))
	return nil, errors.New(string(respbytes))
}

// applyAPIProject configure the envoy using the API project which takes as
// an input in the form of a byte array.
// If the updating envoy fails, it returns an error, if not error would be nil.
func applyAPIProject(payload []byte) error {
	// Readubg the root zip
	zipReader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))

	if err != nil {
		// loggers.LoggerAPI.Errorf("Error occured while unzipping the apictl project. Error: %v", err.Error())
		return err
	}

	// Read the .zip files within the root apis.zip
	for _, file := range zipReader.File {
		// open the zip files
		if strings.HasSuffix(file.Name, ".zip") {
			fmt.Println(file.Name)
			logger.LoggerMgw.Info("Starting zip reading ....", file.Name)

			// Open thezip
			f, err := file.Open()
			logger.LoggerMgw.Info("Error zip reading ....", err)

			if err != nil {
				return err
			}
			defer f.Close()
			//read the files in a each xxxx-api.zip
			r, err := ioutil.ReadAll(f)
			nzip, err := zip.NewReader(bytes.NewReader(r), int64(len(r)))
			for _, fl := range nzip.File {
				fmt.Println(fl.Name)
				logger.LoggerMgw.Info("Starting zip reading inside 1 ....", fl.Name)

				if strings.HasSuffix(fl.Name, "Definitions/swagger.json") {
					logger.LoggerMgw.Info("Starting zip reading inside ....", fl.Name)

					t, err := fl.Open()
					h, err := ioutil.ReadAll(t)
					if err != nil {
						log.Fatal(err)
					}
					logger.LoggerMgw.Info("updating zds ....", t)

					xds.UpdateEnvoy(h)
				}
			}
		}
	}
	return nil
}
