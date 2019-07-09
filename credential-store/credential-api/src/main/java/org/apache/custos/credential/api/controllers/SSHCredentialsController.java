package org.apache.custos.credential.api.controllers;

import org.apache.airavata.custos.credentials.ssh.SSHCredentialEntity;
import org.apache.airavata.custos.vault.VaultManager;
import org.apache.custos.credential.api.resources.SSHCredential;
import org.dozer.DozerBeanMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
public class SSHCredentialsController {

    @Autowired
    private VaultManager vaultManager;

    @Autowired
    private DozerBeanMapper mapper;

    @RequestMapping(value = "/ssh/{gateway}/{token}", method = RequestMethod.GET)
    public SSHCredential getSSHCredential(@PathVariable("gateway") String gateway, @PathVariable("token") String token) throws Exception {
        SSHCredentialEntity credentialEntity = vaultManager.getCredentialEntity(SSHCredentialEntity.class, token, gateway);
        return mapper.map(credentialEntity, SSHCredential.class);
    }

    @RequestMapping(value = "/ssh/{gateway}", method = RequestMethod.POST)
    public String createSSHCredential(@RequestBody SSHCredential sshCredential, @PathVariable("gateway") String gateway) throws Exception {
        SSHCredentialEntity credentialEntity = mapper.map(sshCredential, SSHCredentialEntity.class);
        String token = vaultManager.saveCredentialEntity(credentialEntity, gateway);
        return token;
    }
}
