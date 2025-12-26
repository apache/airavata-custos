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
package org.apache.custos.signer.sdk.ssh;

import com.hierynomus.sshj.key.KeyAlgorithm;
import com.hierynomus.sshj.key.KeyAlgorithms;
import com.hierynomus.sshj.userauth.certificate.Certificate;
import net.i2p.crypto.eddsa.EdDSAPrivateKey;
import net.i2p.crypto.eddsa.spec.EdDSANamedCurveSpec;
import net.i2p.crypto.eddsa.spec.EdDSANamedCurveTable;
import net.i2p.crypto.eddsa.spec.EdDSAPrivateKeySpec;
import net.schmizz.sshj.SSHClient;
import net.schmizz.sshj.common.Buffer;
import net.schmizz.sshj.common.Factory;
import net.schmizz.sshj.common.KeyType;
import net.schmizz.sshj.connection.channel.direct.Session;
import net.schmizz.sshj.sftp.FileAttributes;
import net.schmizz.sshj.sftp.FileMode;
import net.schmizz.sshj.sftp.RemoteResourceInfo;
import net.schmizz.sshj.sftp.SFTPClient;
import net.schmizz.sshj.signature.SignatureRSA;
import net.schmizz.sshj.transport.verification.PromiscuousVerifier;
import net.schmizz.sshj.userauth.keyprovider.KeyProvider;
import net.schmizz.sshj.xfer.FilePermission;
import org.bouncycastle.asn1.pkcs.PrivateKeyInfo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.IOException;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.security.KeyPair;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.interfaces.EdECPrivateKey;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Date;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

/**
 * SSH session implementation using sshj library.
 * Provides high-level SSH operations using certificate-based authentication.
 */
public class SshSession implements AutoCloseable {

    private static final Logger LOGGER = LoggerFactory.getLogger(SshSession.class);

    private final String host;
    private final int port;
    private final String username;
    private final KeyPair keyPair;
    private final byte[] certificate;
    private final Instant validAfter;
    private final Instant validBefore;

    private static final int DEFAULT_CONNECT_TIMEOUT_MS = 30000;
    private static final int DEFAULT_READ_TIMEOUT_MS = 30000;
    private static final int DEFAULT_KEEP_ALIVE_SECONDS = 30;

    private SSHClient sshClient;
    private boolean connected = false;

    public SshSession(String host, int port, String username, KeyPair keyPair,
                      byte[] certificate, Instant validAfter, Instant validBefore) {
        this.host = host;
        this.port = port;
        this.username = username;
        this.keyPair = keyPair;
        this.certificate = certificate;
        this.validAfter = validAfter;
        this.validBefore = validBefore;
    }

    /**
     * Connect to the SSH server
     */
    public void connect() throws IOException {
        if (connected) {
            LOGGER.debug("SSH session already connected");
            return;
        }

        LOGGER.debug("Connecting to SSH server: {}:{} as user: {}", host, port, username);

        // Create SSH client
        sshClient = new SSHClient();

        // Configure client
        sshClient.setConnectTimeout(DEFAULT_CONNECT_TIMEOUT_MS);
        sshClient.setTimeout(DEFAULT_READ_TIMEOUT_MS);
        configureKeyAlgorithms(sshClient);

        configureHostKeyVerification();

        // Connect to server
        sshClient.connect(host, port);

        // Set keep-alive to avoid idle disconnects
        try {
            sshClient.getConnection().getKeepAlive().setKeepAliveInterval(DEFAULT_KEEP_ALIVE_SECONDS);
            sshClient.getConnection().setTimeoutMs(DEFAULT_READ_TIMEOUT_MS);
        } catch (Exception e) {
            LOGGER.debug("Keep-alive configuration not available in this SSHJ version: {}", e.getMessage());
        }

        // Set up certificate authentication
        setupCertificateAuth();

        connected = true;
        LOGGER.info("SSH session connected successfully to {}:{}", host, port);
    }

    /**
     * Execute a command on the remote server
     */
    public CommandResult exec(String command) throws IOException {
        return exec(command, null, 0);
    }

    /**
     * Execute a command with working directory and timeout
     */
    public CommandResult exec(String command, String workingDir, int timeoutSeconds) throws IOException {
        ensureConnected();

        LOGGER.debug("Executing command: {} (workingDir: {}, timeout: {}s)", command, workingDir, timeoutSeconds);

        try (Session session = sshClient.startSession()) {
            Session.Command cmd = session.exec(command);

            // Wait for command completion
            if (timeoutSeconds > 0) {
                cmd.join(timeoutSeconds, TimeUnit.SECONDS);
                if (cmd.getExitStatus() == null) {
                    throw new IOException("Command timed out after " + timeoutSeconds + "s: " + command);
                }
            } else {
                cmd.join();
            }

            String stdout = cmd.getInputStream() != null ? new String(cmd.getInputStream().readAllBytes(), StandardCharsets.UTF_8) : "";
            String stderr = cmd.getErrorStream() != null ? new String(cmd.getErrorStream().readAllBytes(), StandardCharsets.UTF_8) : "";

            int exitCode = cmd.getExitStatus() != null ? cmd.getExitStatus() : -1;
            LOGGER.debug("Command completed with exit code: {}", exitCode);
            return new CommandResult(exitCode, stdout, stderr);

        } catch (Exception e) {
            LOGGER.error("Error executing command: {}", command, e);
            throw new IOException("Failed to execute command: " + command, e);
        }
    }

    /**
     * Upload a file to the remote server
     */
    public void upload(String localPath, String remotePath, int mode) throws IOException {
        ensureConnected();

        LOGGER.debug("Uploading file: {} -> {} (mode: {})", localPath, remotePath, mode);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            sftp.put(localPath, remotePath);

            // Set file permissions if mode is specified
            if (mode > 0) {
                try {
                    FileAttributes attrs = sftp.stat(remotePath);
                    attrs.getPermissions().clear();
                    attrs.getPermissions().addAll(FilePermission.fromMask(mode));
                    sftp.setattr(remotePath, attrs);
                    LOGGER.debug("Set file permissions to {} (octal: {})", mode, Integer.toOctalString(mode));

                } catch (Exception e) {
                    LOGGER.warn("Failed to set file permissions to {}: {}", mode, e.getMessage());
                    // Don't fail the upload if the permission setting fails
                }
            }

            LOGGER.debug("File uploaded successfully");
        }
    }

    /**
     * Download a file from the remote server
     */
    public void download(String remotePath, String localPath) throws IOException {
        ensureConnected();

        LOGGER.debug("Downloading file: {} -> {}", remotePath, localPath);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            sftp.get(remotePath, localPath);
            LOGGER.debug("File downloaded successfully");
        }
    }

    /**
     * Create a directory
     */
    public void mkdir(String path, boolean recursive, int mode) throws IOException {
        ensureConnected();

        LOGGER.debug("Creating directory: {} (recursive: {}, mode: {})", path, recursive, mode);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            if (recursive) {
                sftp.mkdirs(path);
            } else {
                sftp.mkdir(path);
            }

            if (mode > 0) {
                try {
                    FileAttributes attrs = sftp.stat(path);
                    attrs.getPermissions().clear();
                    attrs.getPermissions().addAll(FilePermission.fromMask(mode));
                    sftp.setattr(path, attrs);
                    LOGGER.debug("Set directory permissions to {} (octal: {})", mode, Integer.toOctalString(mode));
                } catch (Exception e) {
                    LOGGER.warn("Failed to set directory permissions to {}: {}", mode, e.getMessage());
                    // Don't fail directory creation if permission setting fails
                }
            }

            LOGGER.debug("Directory created successfully");
        }
    }

    /**
     * Remove a file or directory
     */
    public void rm(String path, boolean recursive) throws IOException {
        ensureConnected();

        LOGGER.debug("Removing: {} (recursive: {})", path, recursive);

        try (net.schmizz.sshj.sftp.SFTPClient sftp = sshClient.newSFTPClient()) {
            if (recursive) {
                sftp.rm(path);
            } else {
                sftp.rm(path);
            }

            LOGGER.debug("Removed successfully");
        }
    }

    /**
     * Move/rename a file
     */
    public void mv(String sourcePath, String destPath) throws IOException {
        ensureConnected();

        LOGGER.debug("Moving: {} -> {}", sourcePath, destPath);

        try (net.schmizz.sshj.sftp.SFTPClient sftp = sshClient.newSFTPClient()) {
            sftp.rename(sourcePath, destPath);
            LOGGER.debug("Moved successfully");
        }
    }

    /**
     * Change file permissions
     */
    public void chmod(String path, int mode) throws IOException {
        ensureConnected();

        LOGGER.debug("Changing permissions: {} -> {}", path, mode);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            FileAttributes attrs = sftp.stat(path);
            attrs.getPermissions().clear();
            attrs.getPermissions().addAll(FilePermission.fromMask(mode));
            sftp.setattr(path, attrs);
            LOGGER.debug("Changed file permissions to {} (octal: {})", mode, Integer.toOctalString(mode));
        }
    }

    /**
     * List directory contents
     */
    public java.util.List<FileInfo> list(String path) throws IOException {
        ensureConnected();

        LOGGER.debug("Listing directory: {}", path);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            java.util.List<RemoteResourceInfo> remoteFiles = sftp.ls(path);

            return remoteFiles.stream()
                    .map(this::toFileInfo)
                    .collect(java.util.stream.Collectors.toList());
        }
    }

    /**
     * Get file information
     */
    public FileInfo stat(String path) throws IOException {
        ensureConnected();

        LOGGER.debug("Getting file info: {}", path);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            FileAttributes attrs = sftp.stat(path);
            // Convert FileAttributes to RemoteResourceInfo for compatibility
            return toFileInfoFromAttributes(path, attrs);
        }
    }

    /**
     * Check if file exists
     */
    public boolean exists(String path) throws IOException {
        try {
            stat(path);
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    /**
     * Read file content as string
     */
    public String read(String path) throws IOException {
        ensureConnected();

        LOGGER.debug("Reading file: {}", path);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            // Use get method instead of read for this SSHJ version
            File tempFile = File.createTempFile("ssh_read", ".tmp");
            try {
                sftp.get(path, tempFile.getAbsolutePath());
                byte[] content = Files.readAllBytes(tempFile.toPath());
                return new String(content);
            } finally {
                tempFile.delete();
            }
        }
    }

    /**
     * Write content to file
     */
    public void write(String content, String remotePath, int mode) throws IOException {
        ensureConnected();

        LOGGER.debug("Writing file: {} (mode: {})", remotePath, mode);

        try (SFTPClient sftp = sshClient.newSFTPClient()) {
            // Use put method instead of write for this SSHJ version
            File tempFile = File.createTempFile("ssh_write", ".tmp");
            try {
                Files.write(tempFile.toPath(), content.getBytes());
                sftp.put(tempFile.getAbsolutePath(), remotePath);
            } finally {
                tempFile.delete();
            }

            if (mode > 0) {
                try {
                    FileAttributes attrs = sftp.stat(remotePath);
                    attrs.getPermissions().clear();
                    attrs.getPermissions().addAll(FilePermission.fromMask(mode));
                    sftp.setattr(remotePath, attrs);
                    LOGGER.debug("Set file permissions to {} (octal: {})", mode, Integer.toOctalString(mode));
                } catch (Exception e) {
                    LOGGER.warn("Failed to set file permissions to {}: {}", mode, e.getMessage());
                    // Don't fail the write if permission setting fails
                }
            }

            LOGGER.debug("File written successfully");
        }
    }

    @Override
    public void close() throws IOException {
        if (sshClient != null && sshClient.isConnected()) {
            LOGGER.debug("Closing SSH session to {}:{}", host, port);
            sshClient.disconnect();
            connected = false;
        } else if (sshClient != null) {
            sshClient.close();
            connected = false;
        }
    }

    private void ensureConnected() throws IOException {
        if (!connected) {
            connect();
        }
    }

    private void setupCertificateAuth() throws IOException {
        Objects.requireNonNull(certificate, "Certificate bytes cannot be null");
        try {
            KeyPair certKeyPair = buildCertificateKeyPair();
            KeyProvider provider = sshClient.loadKeys(certKeyPair);
            sshClient.authPublickey(username, provider);
            LOGGER.debug("Authenticated with certificate for user {}", username);
        } catch (Exception e) {
            throw new IOException("Failed to authenticate with SSH certificate", e);
        }
    }

    private void configureHostKeyVerification() {
        boolean strictHostKey = Boolean.parseBoolean(System.getProperty("custos.ssh.strictHostKey", "false"));
        try {
            sshClient.loadKnownHosts();
            LOGGER.debug("Loaded known_hosts for host key verification");
        } catch (IOException e) {
            if (strictHostKey) {
                throw new RuntimeException("Failed to load known_hosts in strict host key mode", e);
            }
            LOGGER.warn("Unable to load known_hosts, falling back to permissive host key verification: {}", e.getMessage());
        }

        if (!strictHostKey) {
            // Allow unknown hosts in non-strict mode so tests and dev setups succeed
            // TODO allow only when the test profile is active
            sshClient.addHostKeyVerifier(new PromiscuousVerifier());
        }
    }

    private KeyPair buildCertificateKeyPair() throws IOException {
        try {
            Buffer.PlainBuffer buffer = new Buffer.PlainBuffer(certificate);
            PublicKey parsed = buffer.readPublicKey();
            Certificate<?> cert = toCertificate(parsed);
            validateCertificateWindow(cert);
            PrivateKey normalizedPrivate = normalizePrivateKey(cert, keyPair.getPrivate());
            return new KeyPair(parsed, normalizedPrivate);
        } catch (Buffer.BufferException e) {
            throw new IOException("Failed to parse SSH certificate", e);
        }
    }

    @SuppressWarnings("unchecked")
    private Certificate<PublicKey> toCertificate(PublicKey publicKey) {
        if (publicKey instanceof Certificate<?>) {
            return (Certificate<PublicKey>) publicKey;
        }
        throw new IllegalArgumentException("Parsed public key is not an SSH certificate");
    }

    private void validateCertificateWindow(Certificate<?> cert) {
        if (cert == null) {
            return;
        }
        Date now = new Date();
        if (cert.getValidAfter() != null && now.before(cert.getValidAfter())) {
            LOGGER.warn("Certificate not yet valid: validAfter={}, now={}", cert.getValidAfter(), now);
        }
        if (cert.getValidBefore() != null && now.after(cert.getValidBefore())) {
            LOGGER.warn("Certificate expired: validBefore={}, now={}", cert.getValidBefore(), now);
        }
    }

    private PrivateKey normalizePrivateKey(Certificate<?> cert, PrivateKey originalPrivateKey) {
        if (cert == null || originalPrivateKey == null) {
            return originalPrivateKey;
        }
        KeyType certType = KeyType.fromKey(cert);
        if (KeyType.ED25519_CERT.equals(certType)) {
            // SSHJ requires net.i2p EdDSA keys; convert from JDK EdDSA private key if needed
            if (originalPrivateKey instanceof EdDSAPrivateKey) {
                return originalPrivateKey;
            }
            if (originalPrivateKey instanceof EdECPrivateKey) {
                byte[] seed = extractEd25519Seed((EdECPrivateKey) originalPrivateKey);
                if (seed != null) {
                    EdDSANamedCurveSpec spec = EdDSANamedCurveTable.getByName("Ed25519");
                    EdDSAPrivateKeySpec privSpec = new EdDSAPrivateKeySpec(seed, spec);
                    return new EdDSAPrivateKey(privSpec);
                }
            }
        }
        return originalPrivateKey;
    }

    private void configureKeyAlgorithms(SSHClient client) {
        try {
            List<Factory.Named<KeyAlgorithm>> current = new ArrayList<>(client.getTransport().getConfig().getKeyAlgorithms());
            // Prefer RSA SHA2 certificate algorithms if server requires SHA2 for RSA certs
            Factory.Named<KeyAlgorithm> rsaSha256Cert = new KeyAlgorithms.Factory("rsa-sha2-256-cert-v01@openssh.com",
                    new SignatureRSA.FactoryRSASHA256(), KeyType.RSA_CERT);
            Factory.Named<KeyAlgorithm> rsaSha512Cert = new KeyAlgorithms.Factory("rsa-sha2-512-cert-v01@openssh.com",
                    new SignatureRSA.FactoryRSASHA512(), KeyType.RSA_CERT);

            // Prepend to ensure preference
            current.add(0, rsaSha512Cert);
            current.add(0, rsaSha256Cert);
            client.getTransport().getConfig().setKeyAlgorithms(current);
        } catch (Exception e) {
            LOGGER.warn("Unable to configure RSA SHA2 certificate algorithms, continuing with defaults: {}", e.getMessage());
        }
    }

    private byte[] extractEd25519Seed(EdECPrivateKey jdkPriv) {
        try {
            // JDK 17 returns byte[]; newer JDKs may return Optional<byte[]>
            Method m = jdkPriv.getClass().getMethod("getBytes");
            Object result = m.invoke(jdkPriv);
            if (result instanceof byte[]) {
                return (byte[]) result;
            }
            if (result instanceof Optional<?>) {
                Optional<?> opt = (Optional<?>) result;
                Object val = opt.orElse(null);
                if (val instanceof byte[]) {
                    return (byte[]) val;
                }
            }
        } catch (Exception e) {
            LOGGER.debug("Failed to extract Ed25519 seed via reflection: {}", e.getMessage());
        }

        // Fallback: parse PKCS8 to extract seed
        try {
            byte[] encoded = jdkPriv.getEncoded();
            if (encoded != null) {
                byte[] octets = PrivateKeyInfo.getInstance(encoded).getPrivateKey().getOctets();
                if (octets.length == 34 && octets[0] == 0x04 && octets[1] == 0x20) {
                    return Arrays.copyOfRange(octets, 2, 34);
                }
                return octets;
            }
        } catch (Exception e) {
            LOGGER.debug("Failed to extract Ed25519 seed from PKCS8: {}", e.getMessage());
        }
        return null;
    }

    private FileInfo toFileInfo(net.schmizz.sshj.sftp.RemoteResourceInfo info) {
        // Convert FilePermission Set to int mask
        int permissions = FilePermission.toMask(info.getAttributes().getPermissions());

        return new FileInfo(
                info.getName(),
                info.getPath(),
                info.getAttributes().getSize(),
                info.getAttributes().getMtime(),
                permissions,
                info.getAttributes().getType() == FileMode.Type.DIRECTORY
        );
    }

    private FileInfo toFileInfoFromAttributes(String path, FileAttributes attrs) {
        String fileName = java.nio.file.Paths.get(path).getFileName().toString();

        // Convert FilePermission Set to int mask
        int permissions = FilePermission.toMask(attrs.getPermissions());

        return new FileInfo(
                fileName,
                path,
                attrs.getSize(),
                attrs.getMtime(),
                permissions,
                attrs.getType() == FileMode.Type.DIRECTORY
        );
    }

    /**
     * Command execution result
     */
    public static class CommandResult {
        private final int exitCode;
        private final String stdout;
        private final String stderr;

        public CommandResult(int exitCode, String stdout, String stderr) {
            this.exitCode = exitCode;
            this.stdout = stdout;
            this.stderr = stderr;
        }

        public int getExitCode() {
            return exitCode;
        }

        public String getStdout() {
            return stdout;
        }

        public String getStderr() {
            return stderr;
        }

        public boolean isSuccess() {
            return exitCode == 0;
        }
    }

    /**
     * File information container
     */
    public static class FileInfo {
        private final String name;
        private final String path;
        private final long size;
        private final long mtime;
        private final int permissions;
        private final boolean isDirectory;

        public FileInfo(String name, String path, long size, long mtime, int permissions, boolean isDirectory) {
            this.name = name;
            this.path = path;
            this.size = size;
            this.mtime = mtime;
            this.permissions = permissions;
            this.isDirectory = isDirectory;
        }

        public String getName() {
            return name;
        }

        public String getPath() {
            return path;
        }

        public long getSize() {
            return size;
        }

        public long getMtime() {
            return mtime;
        }

        public int getPermissions() {
            return permissions;
        }

        public boolean isDirectory() {
            return isDirectory;
        }
    }
}
