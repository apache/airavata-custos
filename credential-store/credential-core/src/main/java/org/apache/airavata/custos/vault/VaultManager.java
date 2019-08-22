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
package org.apache.airavata.custos.vault;

import com.bettercloud.vault.Vault;
import com.bettercloud.vault.VaultConfig;
import com.bettercloud.vault.VaultException;
import com.bettercloud.vault.response.LogicalResponse;
import org.apache.airavata.custos.credentials.BaseCredentialEntity;
import org.apache.airavata.custos.credentials.ssh.SSHCredentialEntity;
import org.apache.airavata.custos.vault.annotations.VaultPath;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.util.*;

@Service
public class VaultManager {

    private String vaultAddress = "http://127.0.0.1:8200";
    private String vaultToken = "s.AQ3iZYawSgsE4duk8sKs2MGW";

    private Vault vault;

    @PostConstruct
    public void init() throws VaultException {
        final VaultConfig config = new VaultConfig().address(vaultAddress).token(vaultToken).build();
        vault = new Vault(config);
    }

    public <T extends BaseCredentialEntity> T getCredentialEntity(Class<T> clazz, final String token, final String tenant) throws Exception {

        Map<String, String> params = new HashMap<String, String>() {{
            put("token", token);
            put("tenant", tenant);
        }};

        Constructor<T> ctor = clazz.getConstructor();
        T obj = ctor.newInstance();

        for (Class<?> c = clazz; c != null; c = c.getSuperclass()) {
            Field[] fields = c.getDeclaredFields();
            for (Field field : fields) {
                VaultPath vaultPathAnnotation = field.getAnnotation(VaultPath.class);
                if (vaultPathAnnotation != null) {
                    String path = populatePathWithParams(vaultPathAnnotation.path(), params);
                    Map<String,String> data = vault.logical().read(path).getData();
                    field.setAccessible(true);
                    field.set(obj, data.get(vaultPathAnnotation.name()));
                }
            }
        }
        return obj;
    }

    public <T extends BaseCredentialEntity> String saveCredentialEntity(final T credentialEntity, final String tenant) throws Exception {
        final String token = UUID.randomUUID().toString();

        credentialEntity.setGateway(tenant);
        credentialEntity.setToken(token);

        Map<String, String> params = new HashMap<String, String>() {{
            put("token", token);
            put("tenant", tenant);
        }};

        Map<String, Map<String, Object>> summary = new HashMap<>();

        for (Class<?> c = credentialEntity.getClass(); c != null; c = c.getSuperclass()) {
            Field[] fields = c.getDeclaredFields();
            for (Field field : fields) {
                VaultPath vaultPathAnnotation = field.getAnnotation(VaultPath.class);
                if (vaultPathAnnotation != null) {
                    field.setAccessible(true);
                    String vaultPathValue = (String ) field.get(credentialEntity);
                    String path = populatePathWithParams(vaultPathAnnotation.path(), params);
                    Map map = summary.computeIfAbsent(path, (k) -> new HashMap());
                    map.put(vaultPathAnnotation.name(), vaultPathValue);
                }
            }
        }

        for (Map.Entry<String, Map<String, Object>> entry : summary.entrySet()) {
            vault.logical().write(entry.getKey(), entry.getValue());
        }
        return token;
    }

    /**
     * This will resolve parameterized path strings with the provided param map. Example path is like secret/{gateway}/{token}/value.
     * Params map should provide the values for gateway and token parameters.
     *
     * @param path Path with parameters
     * @param params Parameter map
     * @return Resolved path
     * @throws Exception if the path is not in correct format or required parameters are not found in the params map
     */
    private String populatePathWithParams(String path, Map<String, String> params) throws Exception {
        String newPath = "";
        int begin = 0;
        while (true) {
            int startPos = path.indexOf("{", begin);
            if (startPos != -1) {
                int endPos = path.indexOf("}" , begin);
                if (endPos == -1) {
                    throw new Exception("Path " + path + " is not in the correct format");
                } else {
                    newPath += path.substring(begin, startPos);
                    String paramName = path.substring(startPos + 1, endPos);
                    String paramValue = params.get(paramName);
                    if (paramValue == null) {
                        throw new Exception("Parameter can not be found for name " + paramName + " in path " + path);
                    }
                    newPath += paramValue;
                    begin = endPos + 1;
                }
            } else {
                return begin == 0? path: newPath + path.substring(begin);
            }
        }
    }

    public String getVaultAddress() {
        return vaultAddress;
    }

    public void setVaultAddress(String vaultAddress) {
        this.vaultAddress = vaultAddress;
    }

    public String getVaultToken() {
        return vaultToken;
    }

    public void setVaultToken(String vaultToken) {
        this.vaultToken = vaultToken;
    }

    public static void main(String args[]) throws Exception {
        final VaultConfig config = new VaultConfig()
                .address("http://127.0.0.1:8200")
                .token("s.PFc3SbOz1N2wSpV5hWZ56yVI")
                .build();

        final Vault vault = new Vault(config);

        Map<String, String> data = vault.logical().read("secret/hello").getData();
        //System.out.println(data);

        VaultManager manager = new VaultManager();
        manager.init();

        SSHCredentialEntity credentialEntity = new SSHCredentialEntity();
        credentialEntity.setPrivateKey("def");
        credentialEntity.setPublicKey("ddfdfef");
        credentialEntity.setPassphrase("11111");
        String token = manager.saveCredentialEntity(credentialEntity, "seagrid");
        //System.out.println(token);
        SSHCredentialEntity c = manager.getCredentialEntity(SSHCredentialEntity.class, "1a5a8ba8-384b-40ab-8be4-577a1cdb02c3", "seagrid");
    }
}
