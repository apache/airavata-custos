/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.resource.secret.manager.adaptor.outbound;

import com.google.protobuf.GeneratedMessageV3;
import com.jcraft.jsch.JSch;
import com.jcraft.jsch.KeyPair;
import org.apache.commons.io.FileUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


import java.io.File;
import java.util.UUID;


/**
 * This class creates SSH credential from gRPC SSH Credential
 */
public class SSHCredential extends ResourceCredential {

    private static final Logger LOGGER = LoggerFactory.getLogger(SSHCredential.class);

    private String publicKey;
    private String privateKey;
    private String passPhrase;

    public SSHCredential(GeneratedMessageV3 message) throws Exception {
        super(message);
       if (message instanceof org.apache.custos.resource.secret.service.SSHCredential) {

         this.passPhrase =  ((org.apache.custos.resource.secret.service.SSHCredential) message).getPassphrase();
         this.privateKey = ((org.apache.custos.resource.secret.service.SSHCredential) message).getPrivateKey();
         this.publicKey = ((org.apache.custos.resource.secret.service.SSHCredential) message).getPublicKey();

         if (passPhrase == null || passPhrase.trim().equals("")) {
             this.passPhrase = String.valueOf(UUID.randomUUID());
         }
          this.generateKeyPair(this.passPhrase);

       }

    }

    public String getPublicKey() {
        return publicKey;
    }

    public void setPublicKey(String publicKey) {
        this.publicKey = publicKey;
    }

    public String getPrivateKey() {
        return privateKey;
    }

    public void setPrivateKey(String privateKey) {
        this.privateKey = privateKey;
    }

    public String getPassPhrase() {
        return passPhrase;
    }

    public void setPassPhrase(String passPhrase) {
        this.passPhrase = passPhrase;
    }

    private void  generateKeyPair(String passPhrase) throws Exception{
        JSch jsch=new JSch();
        try{
            KeyPair kpair= KeyPair.genKeyPair(jsch, KeyPair.RSA, 2048);
            File file = File.createTempFile("id_rsa", "");
            String fileName = file.getAbsolutePath();

            kpair.writePrivateKey(fileName, passPhrase.getBytes());
            kpair.writePublicKey(fileName + ".pub"  , "");
            kpair.dispose();
            byte[] priKey = FileUtils.readFileToByteArray(new File(fileName));

            byte[] pubKey = FileUtils.readFileToByteArray(new File(fileName + ".pub"));
            this.privateKey = new String(priKey);
            this.publicKey = new String(pubKey);

        }
        catch(Exception e){
            LOGGER.error("Error while creating key pair", e);
            throw new Exception("Error while creating key pair", e);
        }
    }
}
