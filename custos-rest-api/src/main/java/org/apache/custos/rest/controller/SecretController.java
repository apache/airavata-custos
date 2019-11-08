package org.apache.custos.rest.controller;

import org.apache.custos.rest.resources.KVSecretResource;
import org.apache.custos.rest.util.SecretEngine;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/secret")
public class SecretController {

    @Autowired
    private SecretEngine secretEngine;

    /**
     * This saves a single key value pair in vault and returns a vault id
     * @return Unique id for the secret
     */
    @RequestMapping(value = "single", method = RequestMethod.POST)
    public String saveSecret(@RequestBody KVSecretResource secret) {
        return secretEngine.soreKV(Collections.singletonMap(secret.getKey(), secret.getValue()));
    }


    /**
     * This saves multiple key value pairs in vault and returns a single vault id
     * @param secrets
     * @return Unique id for all secrets
     */
    @RequestMapping(value = "multiple", method = RequestMethod.POST)
    public String saveSecrets(@RequestBody List<KVSecretResource> secrets) {
        Map<String, String> secMap = new HashMap<>();
        secrets.forEach(s-> secMap.put(s.getKey(), s.getValue()));
        return secretEngine.soreKV(secMap);
    }

    public List<KVSecretResource> getSecrets(String token) {
        //Optional.ofNullable(secretEngine.getKV(token)).orElse(new HashMap<>());
        //secretEngine.getKV(token).
        return new ArrayList<>();
    }
}
