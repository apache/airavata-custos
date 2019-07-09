package org.apache.airavata.custos.credentials.ssh;

import org.apache.airavata.custos.credentials.BaseCredentialEntity;
import org.apache.airavata.custos.vault.annotations.VaultPath;

public class SSHCredentialEntity extends BaseCredentialEntity {
    @VaultPath(path = "secret/ssh/{gateway}/{token}", name = "private", required = true)
    private String privateKey;

    @VaultPath(path = "secret/ssh/{gateway}/{token}", name = "passphrase")
    private String passphrase;

    @VaultPath(path = "secret/ssh/{gateway}/{token}", name = "public", required = true)
    private String publicKey;

    public String getPrivateKey() {
        return privateKey;
    }

    public void setPrivateKey(String privateKey) {
        this.privateKey = privateKey;
    }

    public String getPassphrase() {
        return passphrase;
    }

    public void setPassphrase(String passphrase) {
        this.passphrase = passphrase;
    }

    public String getPublicKey() {
        return publicKey;
    }

    public void setPublicKey(String publicKey) {
        this.publicKey = publicKey;
    }
}
