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

package org.apache.custos.ssl.certificate.manager;

import org.apache.custos.ssl.certificate.manager.acme.AcmeClient;
import org.apache.custos.ssl.certificate.manager.acme.AcmeConfiguration;
import org.apache.custos.ssl.certificate.manager.custos.CustosClient;
import org.apache.custos.ssl.certificate.manager.custos.CustosConfiguration;
import org.apache.custos.ssl.certificate.manager.nginx.NginxClient;
import org.apache.custos.ssl.certificate.manager.nginx.NginxConfiguration;
import org.quartz.Job;
import org.quartz.JobExecutionContext;
import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.exception.AcmeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.Map;

public class CertUpdater implements Job {

    private static final Logger logger = LoggerFactory.getLogger(CertUpdater.class);

    @Override
    public void execute(JobExecutionContext jobExecutionContext) {
        Configuration config = null;
        String configPath = jobExecutionContext.getJobDetail().getJobDataMap().get(Constants.CONFIG_PATH).toString();

        if (configPath != null && new File(configPath).exists()) {
            try (InputStream in = Files.newInputStream(Paths.get(configPath))) {
                Yaml yaml = new Yaml();
                config = yaml.loadAs(in, Configuration.class);
            } catch (IOException e) {
                logger.error("Error has occurred while reading config:{}", e.getMessage());
            }
        } else {
            config = new Configuration();
            Map<String, String> env = System.getenv();

            // Nginx client configuration
            NginxConfiguration nginxConfiguration = new NginxConfiguration();
            nginxConfiguration.setUrl(env.get(Constants.NGINX_URL));
            nginxConfiguration.setFolderPath(env.get(Constants.NGINX_CHALLENGE_FOLDER_PATH));

            // Acme client configuration
            AcmeConfiguration acmeConfiguration = new AcmeConfiguration();
            acmeConfiguration.setUri(env.get(Constants.ACME_URI));
            acmeConfiguration.setDomains(Arrays.asList(env.get(Constants.ACME_DOMAINS).split(" ")));
            acmeConfiguration.setUserKey(env.get(Constants.ACME_USER_KEY));

            // Custos client configuration
            CustosConfiguration custosConfiguration = new CustosConfiguration();
            custosConfiguration.setHost(env.get(Constants.CUSTOS_HOST));
            custosConfiguration.setPort(Integer.parseInt(env.get(Constants.CUSTOS_PORT)));
            custosConfiguration.setClientId(env.get(Constants.CUSTOS_CLIENT_ID));
            custosConfiguration.setClientSecret(env.get(Constants.CUSTOS_CLIENT_SECRET));

            config.setNginxConfiguration(nginxConfiguration);
            config.setAcmeConfiguration(acmeConfiguration);
            config.setCustosConfiguration(custosConfiguration);
        }

        try {
            AcmeClient acmeClient = new AcmeClient(config.getAcmeConfiguration());
            CustosClient custosClient = new CustosClient(config.getCustosConfiguration());
            NginxClient nginxClient = new NginxClient(config.getNginxConfiguration());

            Order order = acmeClient.getCertificateOrder();
            acmeClient.authorizeDomain(order, nginxClient);
            Certificate certificate = acmeClient.getCertificateCredentials(order);
            String token = custosClient.addCertificate("test", certificate);
            if (token == null || token.isEmpty()) {
                logger.error("Error has occurred while adding certificate to Custos ");
            }
        } catch (AcmeException e) {
            logger.error(e.getMessage());
        } catch (IOException e) {
            logger.error(e.getMessage());
        }
    }
}
