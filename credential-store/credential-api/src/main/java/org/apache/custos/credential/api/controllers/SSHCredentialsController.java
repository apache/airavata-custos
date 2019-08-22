/*
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.credential.api.controllers;

import org.apache.airavata.custos.credentials.ssh.SSHCredentialEntity;
import org.apache.airavata.custos.vault.VaultManager;
import org.apache.custos.credential.api.resources.SSHCredential;
import org.dozer.DozerBeanMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/ssh")
public class SSHCredentialsController {

    @Autowired
    private VaultManager vaultManager;

    @Autowired
    private DozerBeanMapper mapper;

    @RequestMapping(value = "/{tenant}/{token}", method = RequestMethod.GET)
    public SSHCredential getSSHCredential(@PathVariable("tenant") String tenant, @PathVariable("token") String token) throws Exception {
        SSHCredentialEntity credentialEntity = vaultManager.getCredentialEntity(SSHCredentialEntity.class, token, tenant);
        return mapper.map(credentialEntity, SSHCredential.class);
    }

    @RequestMapping(value = "/{tenant}", method = RequestMethod.POST)
    public String createSSHCredential(@RequestBody SSHCredential sshCredential, @PathVariable("tenant") String tenant) throws Exception {
        SSHCredentialEntity credentialEntity = mapper.map(sshCredential, SSHCredentialEntity.class);
        String token = vaultManager.saveCredentialEntity(credentialEntity, tenant);
        return token;
    }
}
