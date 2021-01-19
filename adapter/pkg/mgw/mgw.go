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
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xdsv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	apiServer "github.com/wso2/micro-gw/pkg/api"
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
	"log"
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

func runManagementServer(server xdsv3.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.LoggerMgw.Fatal("failed to listen: ", err)
	}

	// register services
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

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
	srv := xdsv3.NewServer(ctx, cache, nil)

	err := applyAPIProject(fetchAPIs())
	logger.LoggerMgw.Info("SSSSSSSSS", err)
	if err != nil {
		logger.LoggerMgw.Info("Error occured while starting:!!!! ", err)
	}
	logger.LoggerMgw.Info("DDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")
	runManagementServer(srv, port)

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

func fetchAPIs() []byte {
	// URL has to be taken from a config, config toml already has this
	url := "https://172.17.0.1:9443/internal/data/v1/runtime-artifacts"
	logger.LoggerMgw.Info("URL defined to download file ....")

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
	logger.LoggerMgw.Info("Preparing query parameters")

	// Setting authorization header
	req.Header.Set("Authorization", "Basic YWRtaW46YWRtaW4=")
	// Make the request
	logger.LoggerMgw.Info("Starting HTTP client to call API Manager ....")

	resp, err := client.Do(req)
	logger.LoggerMgw.Info("HTTP Client invoked ....")
	logger.LoggerMgw.Info("Response: ", resp)
	logger.LoggerMgw.Info("Error:", err)

	logger.LoggerMgw.Info(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		logger.LoggerMgw.Info("success")
		//err = os.Mkdir("/tmp", 0777)
		if err != nil {
			log.Fatal("Error creating zip archive", err)
		}
		respbytes, err := ioutil.ReadAll(resp.Body)
		err = ioutil.WriteFile("apis.zip", respbytes, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return respbytes

	}
	if err != nil {
		fmt.Println("Error occurred while calling the API", err)
	}
	fmt.Println("body", resp.Body)
	return nil
}

/*
 * Apply the API project to the envoy using xDS APIs
 *
 *
 */
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
			logger.LoggerMgw.Info("Starting Zip file inside apis.zip reading ....", file.Name)

			// Open thezip
			f, err := file.Open()
			logger.LoggerMgw.Info("Error zip reading ....", err)

			if err != nil {
				return err
			}
			defer f.Close()
			//read the files in a each xxxx-api.zip
			r, err := ioutil.ReadAll(f)
			perr := apiServer.ApplyAPIProject(r)
			if err != nil {
				logger.LoggerMgw.Info("Error occurred while applying project", perr)
			}

			// nzip, err := zip.NewReader(bytes.NewReader(r), int64(len(r)))
			// for _, fl := range nzip.File {
			// 	fmt.Println(fl.Name)
			// 	logger.LoggerMgw.Info("Start reading files inside api zip ....", fl.Name)

			// 	if strings.HasSuffix(fl.Name, "Definitions/swagger.json") {
			// 		logger.LoggerMgw.Info("Starting swagger....", fl.Name)

			// 		t, err := fl.Open()
			// 		h, err := ioutil.ReadAll(t)
			// 		if err != nil {
			// 			log.Fatal(err)
			// 		}
			// 		logger.LoggerMgw.Info("updating zds ....", h)

			// 		xds.UpdateEnvoy(h)
			// 	}
			// }
			// 	// loggers.LoggerAPI.Debugf("openAPI file : %v", file.Name)
			// 	unzippedFileBytes, err := readZipFile(file)
			// 	if err != nil {
			// 		loggers.LoggerAPI.Errorf("Error occured while reading the openapi file. %v", err.Error())
			// 		continue
			// 	}
			// 	apiJsn, conversionErr := utills.ToJSON(unzippedFileBytes)
			// 	if conversionErr != nil {
			// 		loggers.LoggerAPI.Errorf("Error converting api file to json: %v", err.Error())
			// 		return conversionErr
			// 	}
			//
		}
	}
	return nil
}
