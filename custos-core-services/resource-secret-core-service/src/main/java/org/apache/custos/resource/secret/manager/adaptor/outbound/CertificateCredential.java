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
import org.apache.commons.codec.binary.Base64;

import java.io.ByteArrayInputStream;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;

/**
 * This class creates CertificateCredential from gRPC CertificateCredential
 */
public class CertificateCredential extends ResourceCredential {

    private X509Certificate certificate;

    private String cert;

    private long lifetime;

    private String notBefore;

    private String notAfter;

    private String privateKey;



    public CertificateCredential(GeneratedMessageV3 message) throws CertificateException {
        super(message);
        if (message instanceof org.apache.custos.resource.secret.service.CertificateCredential) {
            this.privateKey = ((org.apache.custos.resource.secret.service.CertificateCredential) message).getPrivateKey();
            String certi = ((org.apache.custos.resource.secret.service.CertificateCredential) message).getX509Cert();
            Base64 encoder = new Base64(64);
            byte[] decoded = encoder.decode(certi.replaceAll("-----BEGIN CERTIFICATE-----", "").replaceAll("-----END CERTIFICATE-----", ""));
            CertificateFactory cf = CertificateFactory.getInstance("X.509");
            this.certificate = (X509Certificate) cf.generateCertificate(new ByteArrayInputStream(decoded));
            this.notAfter = this.certificate.getNotAfter().toString();
            this.notBefore = this.certificate.getNotBefore().toString();
            this.lifetime = this.certificate.getNotAfter().getTime() - this.certificate.getNotBefore().getTime();
            this.cert = this.certificate.toString();
        }
    }

    public X509Certificate getCertificate() {
        return certificate;
    }


    public void setCertificate(X509Certificate certificate) {
        this.certificate = certificate;
    }

    public String getCert() {
        return cert;
    }

    public void setCert(String cert) {
        this.cert = cert;
    }

    public long getLifetime() {
        return lifetime;
    }

    public void setLifetime(long lifetime) {
        this.lifetime = lifetime;
    }

    public String getNotBefore() {
        return notBefore;
    }

    public void setNotBefore(String notBefore) {
        this.notBefore = notBefore;
    }

    public String getNotAfter() {
        return notAfter;
    }

    public void setNotAfter(String notAfter) {
        this.notAfter = notAfter;
    }

    public String getPrivateKey() {
        return privateKey;
    }

    public void setPrivateKey(String privateKey) {
        this.privateKey = privateKey;
    }
}
