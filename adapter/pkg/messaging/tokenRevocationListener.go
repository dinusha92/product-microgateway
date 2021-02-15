/*
 *  Copyright (c) 2021, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

// Package messaging holds the implementation for event listeners functions
package messaging

import (
	"encoding/json"

	"github.com/streadway/amqp"
	logger "github.com/wso2/micro-gw/loggers"
)

func handleTokenRevocation(deliveries <-chan amqp.Delivery, done chan error) {

	for d := range deliveries {
		var notification EventTokenRevocationNotification
		json.Unmarshal([]byte(string(d.Body)), &notification)
		logger.LoggerMsg.Infof("REVOKED PAYLOAD %+v", notification)
		logger.LoggerMsg.Printf("RevokedToken: %s, Token Type: %s", notification.Event.PayloadData.RevokedToken,
			notification.Event.PayloadData.Type)

		d.Ack(false)
	}
	logger.LoggerMsg.Infof("handle: deliveries channel closed")
	done <- nil
}
