package org.apache.custos.rest.core;

import org.apache.custos.client.profile.service.CustosProfileServiceClientFactory;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.profile.tenant.cpi.exception.TenantProfileServiceException;
import org.apache.custos.rest.util.SecretEngine;
import org.dozer.DozerBeanMapper;
import org.jboss.resteasy.client.jaxrs.ResteasyClientBuilder;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.KeycloakBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.HashMap;
import java.util.Map;

@Configuration
public class AppConfig {

    // Use for future custom mapping scenarios https://www.java-success.com/dozer-with-spring-maven-tutorial/
    @Bean
    public DozerBeanMapper dozerBeanMapper() {
        return new DozerBeanMapper();
    }

    @Bean
    public TenantProfileService.Client tenantClient() throws TenantProfileServiceException {
        return CustosProfileServiceClientFactory.createCustosTenantProfileServiceClient("iam.custos.scigap.org", 8081);
    }

    @Bean
    public SecretEngine secretEngine() {
        return new SecretEngine();
    }

    @Bean Keycloak keycloakAdminClient() {
        ResteasyClientBuilder clientBuilder = new ResteasyClientBuilder()
                .connectionPoolSize(10);

        clientBuilder.disableTrustManager().hostnameVerification(ResteasyClientBuilder.HostnameVerificationPolicy.ANY);
        return KeycloakBuilder.builder()
                .username("admin")
                .realm("master")
                .clientId("admin-cli")
                .password("PASSWORD")
                .resteasyClient(clientBuilder.build())
                .serverUrl("https://iam.custos.scigap.org/auth").build();
    }
}