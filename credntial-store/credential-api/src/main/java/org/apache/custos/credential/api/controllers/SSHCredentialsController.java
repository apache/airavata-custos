package org.apache.custos.credential.api.controllers;

import org.apache.airavata.custos.credentials.ssh.SSHCredentialEntity;
import org.apache.airavata.custos.vault.VaultManager;
import org.apache.custos.credential.api.resources.SSHCredential;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SSHCredentialsController {

    @Autowired
    private VaultManager vaultManager;

    @RequestMapping("/ssh/{gateway}/{token}")
    public SSHCredential getSSHCredential(@PathVariable("gateway") String gateway, @PathVariable("token") String token) throws Exception {
        SSHCredentialEntity credentialEntity = vaultManager.getCredentialEntity(SSHCredentialEntity.class, token, gateway);
        SSHCredential resource = new SSHCredential();
        resource.setPassphrase(credentialEntity.getPassphrase());
        resource.setPrivateKey(credentialEntity.getPrivateKey());
        resource.setPublicKey(credentialEntity.getPublicKey());
        return resource;
    }
}
