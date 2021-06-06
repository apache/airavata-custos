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

import org.apache.custos.ssl.certificate.manager.clients.acme.AcmeClient;
import org.apache.custos.ssl.certificate.manager.configurations.AcmeConfiguration;
import org.apache.custos.ssl.certificate.manager.clients.CustosClient;
import org.apache.custos.ssl.certificate.manager.configurations.CustosConfiguration;
import org.apache.custos.ssl.certificate.manager.clients.NginxClient;
import org.apache.custos.ssl.certificate.manager.configurations.NginxConfiguration;
import org.quartz.Job;
import org.quartz.JobExecutionContext;
import org.shredzone.acme4j.Certificate;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.exception.AcmeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Map;

public class CertUpdater implements Job {

    private static final Logger logger = LoggerFactory.getLogger(CertUpdater.class);

    @Override
    public void execute(JobExecutionContext jobExecutionContext) {
        Configuration config;
        String configPath = jobExecutionContext.getJobDetail().getJobDataMap().get(Constants.CONFIG_PATH).toString();
        try (InputStream in = Files.newInputStream(Paths.get(configPath))) {
            if (configPath != null) {
                Yaml yaml = new Yaml();
                config = yaml.loadAs(in, Configuration.class);
            } else {
                config = new Configuration();
                Map<String, String> env = System.getenv();

                NginxConfiguration nginxConfiguration = new NginxConfiguration(env);
                AcmeConfiguration acmeConfiguration = new AcmeConfiguration(env);
                CustosConfiguration custosConfiguration = new CustosConfiguration(env);

                config.setNginxConfiguration(nginxConfiguration);
                config.setAcmeConfiguration(acmeConfiguration);
                config.setCustosConfiguration(custosConfiguration);
            }

            CustosClient custosClient = new CustosClient(config.getCustosConfiguration());
            NginxClient nginxClient = new NginxClient(config.getNginxConfiguration());
            AcmeClient acmeClient = new AcmeClient(config.getAcmeConfiguration(), custosClient, nginxClient);

            Order order = acmeClient.getCertificateOrder();
            acmeClient.authorizeDomain(order);
            Certificate certificate = acmeClient.getCertificateCredentials(order);
            custosClient.close();
//            String token = custosClient.addCertificate("test", certificate);
//            if (token == null || token.isEmpty()) {
//                logger.error("Error has occurred while adding certificate to Custos ");
//            }
        } catch (AcmeException e) {
            logger.error("Acme Exception : {} ", e.getMessage());
        } catch (IOException e) {
            logger.error("IO Exception: {}", e.getMessage());
        }
    }
}
