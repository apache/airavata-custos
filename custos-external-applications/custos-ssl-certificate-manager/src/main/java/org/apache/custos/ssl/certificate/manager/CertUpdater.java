package org.apache.custos.ssl.certificate.manager;

import org.apache.custos.ssl.certificate.manager.acme.AcmeClient;
import org.apache.custos.ssl.certificate.manager.acme.AcmeConfiguration;
import org.apache.custos.ssl.certificate.manager.nginx.NginxClient;
import org.apache.custos.ssl.certificate.manager.nginx.NginxConfiguration;
import org.quartz.Job;
import org.quartz.JobExecutionContext;
import org.shredzone.acme4j.Order;
import org.shredzone.acme4j.exception.AcmeException;
import org.yaml.snakeyaml.Yaml;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.Map;

public class CertUpdater implements Job {

    @Override
    public void execute(JobExecutionContext jobExecutionContext) {
        Configuration config = null;
        if (new File("config.yaml").exists()) {
            try (InputStream in = Files.newInputStream(Paths.get("config.yaml"))) {
                Yaml yaml = new Yaml();
                config = yaml.loadAs(in, Configuration.class);
            } catch (IOException e) {
                e.printStackTrace();
            }
        } else {
            config = new Configuration();
            Map<String, String> env = System.getenv();

            NginxConfiguration nginxConfiguration = new NginxConfiguration();
            nginxConfiguration.setUrl(env.get("NGINX_URL"));
            nginxConfiguration.setFolderPath(env.get("NGINX_CHALLENGE_FOLDER_PATH"));

            AcmeConfiguration acmeConfiguration = new AcmeConfiguration();
            acmeConfiguration.setSessionUri(env.get("ACME_SESSION_URI"));
            acmeConfiguration.setDomains(Arrays.asList(env.get("ACME_DOMAINS").split(" ")));

            config.setNginxConfiguration(nginxConfiguration);
            config.setAcmeConfiguration(acmeConfiguration);
        }

        try {
            AcmeClient acmeClient = new AcmeClient(config.getAcmeConfiguration());
            NginxClient nginxClient = new NginxClient(config.getNginxConfiguration());
            Order order = acmeClient.getCertificateOrder();
            acmeClient.authorizeDomain(order, nginxClient);
            acmeClient.getCertificates(order);
        } catch (AcmeException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
