package org.apache.airavata.custos.credentials;

public class BaseCredentialEntity {
    private String token;
    private String gateway;

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }

    public String getGateway() {
        return gateway;
    }

    public void setGateway(String gateway) {
        this.gateway = gateway;
    }
}
