package org.apache.airavata.custos.credentials.aws;

import org.apache.airavata.custos.credentials.BaseCredentialEntity;
import org.apache.airavata.custos.vault.annotations.VaultPath;

public class AWSCredentialEntity extends BaseCredentialEntity {

    @VaultPath(path = "secret/aws/{gateway}/{token}", name = "access_key_id", required = true)
    private String accessKey;

    @VaultPath(path = "secret/aws/{gateway}/{token}", name = "secret_access_key")
    private String secretKey;

    public String getAccessKey() {
        return accessKey;
    }

    public void setAccessKey(String accessKey) {
        this.accessKey = accessKey;
    }

    public String getSecretKey() {
        return secretKey;
    }

    public void setSecretKey(String secretKey) {
        this.secretKey = secretKey;
    }
}
