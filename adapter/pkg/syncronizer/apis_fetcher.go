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
 */

/*
 * Package "synchronizer" contains artifacts relate to fetching APIs and
 * API related updates from the API manager event-hub.
 * This file contains operates to retrieve APIs and API updates.
 */

package synchronizer

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	xds "github.com/wso2/micro-gw/pkg/xds"
)

//do the API call
// There is 2 ways API call should be done
// 1.

//

// func loadAPIs() (bool, error) {

// 	url := "https://localhost:9443/internal/data/v1/runtime-artifacts?gatewayLabel=Production%20and%20Sandbox&type=Envoy"
// 	response, err := http.Get(url)
// 	if err != nil {
// 		fmt.Printf("Error occurred while calling the API")
// 	}

// 	defer response.Body.Close()

// 	return false, fmt.Errorf("error occurred while loading APIs")
// }

/*
 * This method will fetch the APIs from the API Manager
 * Todo: As of now, it will retireve all the APIs from the API Manager
 * It also should have the capability to fetch a single API upon update
 * event
 *
 */
func fetchAPIs() []byte {
	// URL has to be taken from a config, config toml already has this
	url := "https://localhost:9443/internal/data/v1/runtime-artifacts"

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
	// Setting authorization header
	req.Header.Set("Authorization", "Basic YWRtaW46YWRtaW4=")
	// Make the request
	resp, err := client.Do(req)
	fmt.Println(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		fmt.Println("success")
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
			// Open thezip
			f, err := file.Open()

			if err != nil {
				return err
			}
			defer f.Close()
			//read the files in a each xxxx-api.zip
			r, err := ioutil.ReadAll(f)
			nzip, err := zip.NewReader(bytes.NewReader(r), int64(len(r)))
			for _, fl := range nzip.File {
				fmt.Println(fl.Name)
				if strings.HasSuffix(file.Name, "Definitions/swagger.json") {
					r, err := ioutil.ReadAll(f)
					if err != nil {
						log.Fatal(err)
					}
					xds.UpdateEnvoy(r)
				}
			}
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