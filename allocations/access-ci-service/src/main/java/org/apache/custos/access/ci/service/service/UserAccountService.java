/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */
package org.apache.custos.access.ci.service.service;

import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.PersonEntity;
import org.apache.custos.access.ci.service.repo.ClusterAccountRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

/**
 * Service for provisioning and managing Cluster Accounts (usernames).
 */
@Service
public class UserAccountService {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserAccountService.class);

    private final ClusterAccountRepository clusterAccountRepository;

    public UserAccountService(ClusterAccountRepository clusterAccountRepository) {
        this.clusterAccountRepository = clusterAccountRepository;
    }

    /**
     * Provisions a new, unique cluster account for a given person.
     *
     * @param person The PersonEntity that the user account should be created
     * @return The newly created and saved ClusterAccountEntity
     */
    @Transactional
    public ClusterAccountEntity provisionClusterAccount(PersonEntity person) {
        // TODO Replace with external source of truth (e.g., COmanage) lookup for PersonID and username

        String proposedUsername = (person.getFirstName().trim().charAt(0) + person.getLastName().trim().replace(" ", "-")).toLowerCase();
        String uniqueUsername = ensureUniqueUsername(proposedUsername);

        LOGGER.info("Provisioning new cluster account with username [{}] for person [{}]", uniqueUsername, person.getId());

        ClusterAccountEntity newClusterAccount = new ClusterAccountEntity();
        newClusterAccount.setId(UUID.randomUUID().toString());
        newClusterAccount.setPerson(person);
        newClusterAccount.setUsername(uniqueUsername);
        clusterAccountRepository.save(newClusterAccount);

        return newClusterAccount;
    }

    private String ensureUniqueUsername(String baseUsername) {
        String candidate = baseUsername;
        int suffix = 0;
        while (clusterAccountRepository.findByUsername(candidate).isPresent()) {
            suffix++;
            candidate = baseUsername + suffix;
        }
        if (suffix > 0) {
            LOGGER.warn("Base username '{}' was already taken. Generated unique username '{}'.", baseUsername, candidate);
        }
        return candidate;
    }
}

