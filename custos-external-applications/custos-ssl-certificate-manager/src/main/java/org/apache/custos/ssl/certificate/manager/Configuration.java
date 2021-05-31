package org.apache.custos.ssl.certificate.manager;

import org.apache.custos.ssl.certificate.manager.acme.AcmeConfiguration;
import org.apache.custos.ssl.certificate.manager.nginx.NginxConfiguration;

public final class Configuration {

    private AcmeConfiguration acmeConfiguration;
    private NginxConfiguration nginxConfiguration;

    public AcmeConfiguration getAcmeConfiguration() {
        return acmeConfiguration;
    }
    public NginxConfiguration getNginxConfiguration() {
        return nginxConfiguration;
    }

    public void setAcmeConfiguration(AcmeConfiguration acmeConfiguration) {
        this.acmeConfiguration = acmeConfiguration;
    }

    public void setNginxConfiguration(NginxConfiguration nginxConfiguration) {
        this.nginxConfiguration = nginxConfiguration;
    }
}
