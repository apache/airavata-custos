/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.ssl.certificate.manager.clients;

import org.apache.custos.ssl.certificate.manager.clients.utils.QueryString;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;

public class NginxClient {
    private static final Logger logger = LoggerFactory.getLogger(NginxClient.class);

    public static class RequestBuilder {
        private String url;
        private String method;
        private String fileName;
        private String fileContent;

        public RequestBuilder(String url) {
            this.url = url;
        }

        public RequestBuilder setHttpMethod(String method) {
            this.method = method;
            return this;
        }

        public RequestBuilder setFileName(String fileName) {
            this.fileName = fileName;
            return this;
        }

        public RequestBuilder setFileContent(String fileContent) {
            this.fileContent = fileContent;
            return this;
        }

        public boolean send() {
            QueryString query = null;
            if (this.fileName != null) {
                query = new QueryString("file", this.fileName);
            }

            if (this.fileContent != null) {
                if (query != null) {
                    query.add("content", this.fileContent);
                } else {
                    query = new QueryString("content", this.fileContent);
                }
            }

            HttpURLConnection con = null;
            int status = 0;
            try {
                URL url = new URL(this.url + "?" + query);
                con = (HttpURLConnection) url.openConnection();
                con.setRequestMethod(this.method);
                status = con.getResponseCode();
            } catch (IOException e) {
                logger.info("Error in nginx client: {}", e.getMessage());
            } finally {
                con.disconnect();
            }

            return status == HttpURLConnection.HTTP_OK ? true : false;
        }
    }
}
