/*
 * Copyright (c) 2021, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */


package org.wso2am.micro.gw.tests.controlplane;

import java.io.File;
import java.util.HashMap;
import java.util.Map;

import com.github.dockerjava.zerodep.shaded.org.apache.hc.core5.http.HttpStatus;
import io.netty.handler.codec.http.HttpHeaderNames;
import org.testng.Assert;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Test;
import org.wso2am.micro.gw.tests.common.BaseTestCase;
import org.wso2am.micro.gw.tests.util.ApiProjectGenerator;
import org.wso2am.micro.gw.tests.util.HttpClientRequest;
import org.wso2am.micro.gw.tests.util.HttpResponse;
import org.wso2am.micro.gw.tests.util.HttpsPostMultipart;
import org.wso2am.micro.gw.tests.util.TestConstant;
import org.wso2.am.integration.test.impl.*;

public class ControlPlaneTest  extends BaseTestCase {
    @BeforeClass(description = "initialise the setup")
    void start() throws Exception {
        super.startAPIMWithMGW(null,true);

    }

    @AfterClass(description = "stop the setup")
    void stop() {
        super.stopAPIMWithMGW();
    }


    @Test(description = "Test to check the api deployment is working")
    public void apiDeployTest() throws Exception {

        //api yaml file should put to the resources/apis/openApis folder
        HttpResponse response = null;
        boolean exit = true;
        do {
            response = HttpClientRequest.doGet("http://localhost:9763/services/Version");
            if (response != null) {
                System.out.println("Response:" + response.getResponseCode());
                 if ( response.getResponseCode()== 200) {
                    exit = false;
                }
            }
        } while (exit);
        restAPIPublisher
        Assert.assertNotNull(response);
        Assert.assertEquals(response.getResponseCode(), HttpStatus.SC_OK,"Response code mismatched");
    }

}