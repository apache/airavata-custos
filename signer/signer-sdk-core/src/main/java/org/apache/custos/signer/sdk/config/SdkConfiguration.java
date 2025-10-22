/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the specific language
 * governing permissions and limitations under the License.
 *
 */
package org.apache.custos.signer.sdk.config;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * SDK configuration for Custos SSH operations.
 * Loads configuration from YAML or programmatic builder.
 */
public class SdkConfiguration {

    private static final Logger logger = LoggerFactory.getLogger(SdkConfiguration.class);

    @JsonProperty("sdk")
    private SdkConfig sdk;

    public SdkConfiguration() {
    }

    public SdkConfiguration(SdkConfig sdk) {
        this.sdk = sdk;
    }

    public SdkConfig getSdk() {
        return sdk;
    }

    public void setSdk(SdkConfig sdk) {
        this.sdk = sdk;
    }

    /**
     * Get tenant ID (inherited by all clients)
     */
    public String getTenantId() {
        return sdk != null ? sdk.getTenantId() : null;
    }

    /**
     * Get client configuration by alias
     */
    public Optional<ClientConfig> getClientConfig(String alias) {
        if (sdk == null || sdk.getClients() == null) {
            return Optional.empty();
        }

        return sdk.getClients().stream()
                .filter(client -> alias.equals(client.getAlias()))
                .findFirst();
    }

    /**
     * Get signer service address
     */
    public String getSignerServiceAddress() {
        return sdk != null && sdk.getSigner() != null ?
                sdk.getSigner().getAddress() : "localhost:9095";
    }

    /**
     * Check if TLS is enabled
     */
    public boolean isTlsEnabled() {
        return sdk != null && sdk.getSigner() != null &&
                sdk.getSigner().getTls() != null && sdk.getSigner().getTls().isEnabled();
    }

    /**
     * Get key store backend type
     */
    public String getKeyStoreBackend() {
        return sdk != null && sdk.getKeyStore() != null ?
                sdk.getKeyStore().getBackend() : "in-memory";
    }

    /**
     * Main SDK configuration
     */
    public static class SdkConfig {
        @JsonProperty("tenant-id")
        private String tenantId;

        @JsonProperty("signer")
        private SignerConfig signer;

        @JsonProperty("keystore")
        private KeyStoreConfig keyStore;

        @JsonProperty("clients")
        private List<ClientConfig> clients;

        public String getTenantId() {
            return tenantId;
        }

        public void setTenantId(String tenantId) {
            this.tenantId = tenantId;
        }

        public SignerConfig getSigner() {
            return signer;
        }

        public void setSigner(SignerConfig signer) {
            this.signer = signer;
        }

        public KeyStoreConfig getKeyStore() {
            return keyStore;
        }

        public void setKeyStore(KeyStoreConfig keyStore) {
            this.keyStore = keyStore;
        }

        public List<ClientConfig> getClients() {
            return clients;
        }

        public void setClients(List<ClientConfig> clients) {
            this.clients = clients;
        }
    }

    /**
     * Signer service configuration
     */
    public static class SignerConfig {
        @JsonProperty("address")
        private String address = "localhost:9095";

        @JsonProperty("tls")
        private TlsConfig tls;

        public String getAddress() {
            return address;
        }

        public void setAddress(String address) {
            this.address = address;
        }

        public TlsConfig getTls() {
            return tls;
        }

        public void setTls(TlsConfig tls) {
            this.tls = tls;
        }
    }

    /**
     * TLS configuration
     */
    public static class TlsConfig {
        @JsonProperty("enabled")
        private boolean enabled = false;

        @JsonProperty("trust-store")
        private String trustStore;

        @JsonProperty("trust-store-password")
        private String trustStorePassword;

        public boolean isEnabled() {
            return enabled;
        }

        public void setEnabled(boolean enabled) {
            this.enabled = enabled;
        }

        public String getTrustStore() {
            return trustStore;
        }

        public void setTrustStore(String trustStore) {
            this.trustStore = trustStore;
        }

        public String getTrustStorePassword() {
            return trustStorePassword;
        }

        public void setTrustStorePassword(String trustStorePassword) {
            this.trustStorePassword = trustStorePassword;
        }
    }

    /**
     * Key store configuration
     */
    public static class KeyStoreConfig {
        @JsonProperty("backend")
        private String backend = "in-memory";

        @JsonProperty("vault")
        private VaultKeyStoreConfig vault;

        @JsonProperty("database")
        private DatabaseKeyStoreConfig database;

        public String getBackend() {
            return backend;
        }

        public void setBackend(String backend) {
            this.backend = backend;
        }

        public VaultKeyStoreConfig getVault() {
            return vault;
        }

        public void setVault(VaultKeyStoreConfig vault) {
            this.vault = vault;
        }

        public DatabaseKeyStoreConfig getDatabase() {
            return database;
        }

        public void setDatabase(DatabaseKeyStoreConfig database) {
            this.database = database;
        }
    }

    /**
     * Vault key store configuration
     */
    public static class VaultKeyStoreConfig {
        @JsonProperty("address")
        private String address;

        @JsonProperty("token")
        private String token;

        public String getAddress() {
            return address;
        }

        public void setAddress(String address) {
            this.address = address;
        }

        public String getToken() {
            return token;
        }

        public void setToken(String token) {
            this.token = token;
        }
    }

    /**
     * Database key store configuration
     */
    public static class DatabaseKeyStoreConfig {
        @JsonProperty("url")
        private String url;

        @JsonProperty("username")
        private String username;

        @JsonProperty("password")
        private String password;

        public String getUrl() {
            return url;
        }

        public void setUrl(String url) {
            this.url = url;
        }

        public String getUsername() {
            return username;
        }

        public void setUsername(String username) {
            this.username = username;
        }

        public String getPassword() {
            return password;
        }

        public void setPassword(String password) {
            this.password = password;
        }
    }

    /**
     * Client configuration
     */
    public static class ClientConfig {
        @JsonProperty("alias")
        private String alias;

        @JsonProperty("client-id")
        private String clientId;

        @JsonProperty("client-secret")
        private String clientSecret;

        public ClientConfig() {
        }

        public ClientConfig(String alias, String clientId, String clientSecret) {
            this.alias = alias;
            this.clientId = clientId;
            this.clientSecret = clientSecret;
        }

        public String getAlias() {
            return alias;
        }

        public void setAlias(String alias) {
            this.alias = alias;
        }

        public String getClientId() {
            return clientId;
        }

        public void setClientId(String clientId) {
            this.clientId = clientId;
        }

        public String getClientSecret() {
            return clientSecret;
        }

        public void setClientSecret(String clientSecret) {
            this.clientSecret = clientSecret;
        }
    }

    /**
     * Builder for programmatic configuration
     */
    public static class Builder {
        private final SdkConfiguration config = new SdkConfiguration();
        private final SdkConfig sdkConfig = new SdkConfig();
        private final SignerConfig signerConfig = new SignerConfig();
        private final KeyStoreConfig keyStoreConfig = new KeyStoreConfig();
        private final Map<String, ClientConfig> clients = new HashMap<>();

        public Builder signerServiceAddress(String address) {
            signerConfig.setAddress(address);
            return this;
        }

        public Builder tlsEnabled(boolean enabled) {
            if (signerConfig.getTls() == null) {
                signerConfig.setTls(new TlsConfig());
            }
            signerConfig.getTls().setEnabled(enabled);
            return this;
        }

        public Builder trustStore(String trustStore, String password) {
            if (signerConfig.getTls() == null) {
                signerConfig.setTls(new TlsConfig());
            }
            signerConfig.getTls().setTrustStore(trustStore);
            signerConfig.getTls().setTrustStorePassword(password);
            return this;
        }

        public Builder keyStoreBackend(String backend) {
            keyStoreConfig.setBackend(backend);
            return this;
        }

        public Builder vaultKeyStore(String address, String token) {
            VaultKeyStoreConfig vaultConfig = new VaultKeyStoreConfig();
            vaultConfig.setAddress(address);
            vaultConfig.setToken(token);
            keyStoreConfig.setVault(vaultConfig);
            return this;
        }

        public Builder databaseKeyStore(String url, String username, String password) {
            DatabaseKeyStoreConfig dbConfig = new DatabaseKeyStoreConfig();
            dbConfig.setUrl(url);
            dbConfig.setUsername(username);
            dbConfig.setPassword(password);
            keyStoreConfig.setDatabase(dbConfig);
            return this;
        }

        public Builder tenantId(String tenantId) {
            sdkConfig.setTenantId(tenantId);
            return this;
        }

        public Builder addClient(String alias, String clientId, String clientSecret) {
            clients.put(alias, new ClientConfig(alias, clientId, clientSecret));
            return this;
        }

        public SdkConfiguration build() {
            sdkConfig.setSigner(signerConfig);
            sdkConfig.setKeyStore(keyStoreConfig);
            sdkConfig.setClients(List.copyOf(clients.values()));
            config.setSdk(sdkConfig);
            return config;
        }
    }
}
