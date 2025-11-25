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
package org.apache.custos.signer.sdk;

import java.security.KeyPair;
import java.time.Instant;

/**
 * Immutable response containing all SSH connection materials.
 * <p>
 * This class provides everything needed to establish an SSH connection:
 * - Private and public keys (both as objects and string formats)
 * - Signed certificate (both as bytes and OpenSSH string format)
 * - Connection metadata (target host, port, username)
 * - Certificate metadata (serial, validity, CA fingerprint)
 * <p>
 *
 * @param keyPair          In-memory keypair object
 * @param privateKeyPem    Private key in PEM format
 * @param publicKeyOpenSsh Public key in OpenSSH format
 * @param opensshCert      Certificate in OpenSSH string format
 * @param certBytes        Certificate bytes (defensively copied in and out)
 * @param serial           Certificate serial number
 * @param validAfter       Certificate validity start
 * @param validBefore      Certificate validity end
 * @param caFingerprint    CA fingerprint
 * @param targetHost       Target SSH host
 * @param targetPort       Target SSH port
 * @param targetUsername   Target SSH username
 */
public record CertificateMaterials(KeyPair keyPair, String privateKeyPem, String publicKeyOpenSsh, String opensshCert,
                                   byte[] certBytes, long serial, Instant validAfter, Instant validBefore,
                                   String caFingerprint, String targetHost, int targetPort, String targetUsername) {

    public CertificateMaterials {
        // Defensive copy of certBytes
        certBytes = (certBytes != null) ? certBytes.clone() : null;
    }

    @Override
    public byte[] certBytes() {
        return certBytes != null ? certBytes.clone() : null;
    }
}

