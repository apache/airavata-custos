package org.apache.custos.credential.api.controllers;

import org.apache.airavata.custos.credentials.aws.AWSCredentialEntity;
import org.apache.airavata.custos.credentials.ssh.SSHCredentialEntity;
import org.apache.airavata.custos.vault.VaultManager;
import org.apache.custos.credential.api.resources.AWSCredntial;
import org.apache.custos.credential.api.resources.SSHCredential;
import org.dozer.DozerBeanMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/aws")
public class AWSCredentialController {
    @Autowired
    private VaultManager vaultManager;

    @Autowired
    private DozerBeanMapper mapper;

    @RequestMapping(value = "/{gateway}/{token}", method = RequestMethod.GET)
    public AWSCredntial getAWSCredential(@PathVariable("gateway") String gateway, @PathVariable("token") String token) throws Exception {
        AWSCredentialEntity credentialEntity = vaultManager.getCredentialEntity(AWSCredentialEntity.class, token, gateway);
        return mapper.map(credentialEntity, AWSCredntial.class);
    }

    @RequestMapping(value = "/{gateway}", method = RequestMethod.POST)
    public String createAWSCredential(@RequestBody AWSCredntial credntial, @PathVariable("gateway") String gateway) throws Exception {
        AWSCredentialEntity credentialEntity = mapper.map(credntial, AWSCredentialEntity.class);
        String token = vaultManager.saveCredentialEntity(credentialEntity, gateway);
        return token;
    }
}
