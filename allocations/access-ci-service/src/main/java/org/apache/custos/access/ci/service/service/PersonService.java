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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.access.ci.service.service;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.PersonDnsEntity;
import org.apache.custos.access.ci.service.model.PersonEntity;
import org.apache.custos.access.ci.service.repo.ClusterAccountRepository;
import org.apache.custos.access.ci.service.repo.PersonDnsRepository;
import org.apache.custos.access.ci.service.repo.PersonRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Service responsible for managing the lifecycle of Person entities and their associated DNs.
 */
@Service
public class PersonService {

    private static final Logger LOGGER = LoggerFactory.getLogger(PersonService.class);

    private final PersonRepository personRepository;
    private final PersonDnsRepository personDnsRepository;
    private final ClusterAccountRepository clusterAccountRepository;

    public PersonService(PersonRepository personRepository, PersonDnsRepository personDnsRepository, ClusterAccountRepository clusterAccountRepository) {
        this.personRepository = personRepository;
        this.personDnsRepository = personDnsRepository;
        this.clusterAccountRepository = clusterAccountRepository;
    }

    /**
     * Finds a Person by their AMIE Global ID or creates a new one if not found.
     *
     * @param packetBody The "body" of the AMIE packet containing user details.
     * @return The existing or newly created PersonEntity.
     */
    @Transactional
    public PersonEntity findOrCreatePersonFromPacket(JsonNode packetBody) {
        String accessGlobalId = packetBody.path("UserGlobalID").asText();
        Assert.hasText(accessGlobalId, "Packet body must contain a 'UserGlobalID'.");

        Optional<PersonEntity> existingPerson = personRepository.findByAccessGlobalId(accessGlobalId);
        if (existingPerson.isPresent()) {
            LOGGER.info("Found existing person with local ID [{}] for access_global_id [{}]", existingPerson.get().getId(), accessGlobalId);
            return existingPerson.get();
        }

        // If not found, create a new person
        LOGGER.info("No person found for access_global_id [{}]. Creating a new person record.", accessGlobalId);
        PersonEntity newPerson = new PersonEntity();
        newPerson.setId(UUID.randomUUID().toString());
        newPerson.setAccessGlobalId(accessGlobalId);
        newPerson.setFirstName(packetBody.path("UserFirstName").asText(""));
        newPerson.setLastName(packetBody.path("UserLastName").asText(""));
        newPerson.setEmail(packetBody.path("UserEmail").asText(""));
        newPerson.setOrganization(packetBody.path("UserOrganization").asText(null));
        newPerson.setOrgCode(packetBody.path("UserOrgCode").asText(null));
        newPerson.setNsfStatusCode(packetBody.path("NsfStatusCode").asText(null));
        personRepository.save(newPerson);

        // Save their associated DNs
        JsonNode dnList = packetBody.path("UserDnList");
        if (dnList.isArray()) {
            for (JsonNode dnNode : dnList) {
                String dn = dnNode.asText(null);
                if (dn != null && !dn.isBlank()) {
                    PersonDnsEntity pde = new PersonDnsEntity();
                    pde.setPerson(newPerson);
                    pde.setDn(dn);
                    personDnsRepository.save(pde);
                }
            }
        }
        return newPerson;
    }


    @Transactional
    public void replaceFromModifyPacket(JsonNode body) {
        String personId = body.path("PersonID").asText(null);
        Assert.hasText(personId, "Missing required 'PersonID' in request_user_modify replace body");

        PersonEntity person = personRepository.findById(personId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown local PersonID: " + personId));

        if (body.has("UserFirstName")) person.setFirstName(body.path("UserFirstName").asText(person.getFirstName()));
        if (body.has("UserLastName")) person.setLastName(body.path("UserLastName").asText(person.getLastName()));
        if (body.has("UserEmail")) person.setEmail(body.path("UserEmail").asText(person.getEmail()));
        if (body.has("UserOrganization")) person.setOrganization(body.path("UserOrganization").asText(null));
        if (body.has("UserOrgCode")) person.setOrgCode(body.path("UserOrgCode").asText(null));
        if (body.has("NsfStatusCode")) person.setNsfStatusCode(body.path("NsfStatusCode").asText(null));
        personRepository.save(person);

        Set<String> newDns = new HashSet<>();
        JsonNode dnList = body.path("UserDnList");
        if (dnList != null && dnList.isArray()) {
            for (JsonNode dnNode : dnList) {
                String dn = dnNode.asText(null);
                if (dn != null && !dn.isBlank()) newDns.add(dn);
            }
        }

        if (newDns.isEmpty()) {
            if (body.has("UserDnList")) {
                personDnsRepository.deleteByPerson_Id(personId);
            }
        } else {
            personDnsRepository.deleteByPerson_IdAndDnNotIn(personId, new ArrayList<>(newDns));
            for (String dn : newDns) {
                if (!personDnsRepository.existsByPerson_IdAndDn(personId, dn)) {
                    PersonDnsEntity p = new PersonDnsEntity();
                    p.setPerson(person);
                    p.setDn(dn);
                    personDnsRepository.save(p);
                }
            }
        }
    }

    @Transactional
    public void persistDnsForPerson(String personId, JsonNode dnList) {
        Assert.hasText(personId, "personId must not be blank");

        if (dnList == null || !dnList.isArray() || dnList.isEmpty()) {
            return;
        }

        PersonEntity person = personRepository.findById(personId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown local PersonID: " + personId));

        Set<String> incomingDns = new HashSet<>();
        for (JsonNode dnNode : dnList) {
            String dn = dnNode.asText(null);
            if (dn != null && !dn.isBlank()) {
                incomingDns.add(dn);
            }
        }

        if (incomingDns.isEmpty()) {
            return;
        }

        for (String dn : incomingDns) {
            if (!personDnsRepository.existsByPerson_IdAndDn(personId, dn)) {
                PersonDnsEntity pde = new PersonDnsEntity();
                pde.setPerson(person);
                pde.setDn(dn);
                try {
                    personDnsRepository.save(pde);
                    LOGGER.info("Persisted new DN for person [{}].", personId);
                    LOGGER.debug("Persisted DN [{}] for person [{}].", dn, personId);
                } catch (DataIntegrityViolationException ex) {
                    LOGGER.debug("DN already exists for person [{}] (concurrent insert), skipping.", personId);
                }
            } else {
                LOGGER.debug("DN [{}] already exists for person [{}], skipping.", dn, personId);
            }
        }
    }

    @Transactional
    public void deleteFromModifyPacket(JsonNode body) {
        String personId = body.path("PersonID").asText(null);
        Assert.hasText(personId, "Missing required 'PersonID' in request_user_modify delete body");

        // Cascades will remove DNs and cluster accounts
        personRepository.deleteById(personId);
    }

    @Transactional
    public void mergePersons(String survivingPersonId, String retiringPersonId) {
        LOGGER.info("Merging person {} into {}", retiringPersonId, survivingPersonId);

        PersonEntity survivingPerson = personRepository.findById(survivingPersonId)
                .orElseThrow(() -> new IllegalStateException("Could not find surviving person with local ID: " + survivingPersonId));

        PersonEntity retiringPerson = personRepository.findById(retiringPersonId)
                .orElseThrow(() -> new IllegalStateException("Could not find retiring person with local ID: " + retiringPersonId));

        // Re-associate all cluster accounts
        for (ClusterAccountEntity account : retiringPerson.getClusterAccounts()) {
            LOGGER.info("Moving cluster account '{}' from retiring person to surviving person", account.getUsername());
            account.setPerson(survivingPerson);
            clusterAccountRepository.save(account);
        }

        // Merge DNs, avoiding duplicates
        Set<String> survivingDns = survivingPerson.getDnsEntries().stream()
                .map(PersonDnsEntity::getDn)
                .collect(Collectors.toSet());

        for (PersonDnsEntity retiringDn : retiringPerson.getDnsEntries()) {
            if (!survivingDns.contains(retiringDn.getDn())) {
                LOGGER.info("Moving DN '{}' from retiring person to surviving person", retiringDn.getDn());
                retiringDn.setPerson(survivingPerson);
                personDnsRepository.save(retiringDn);
            } else {
                personDnsRepository.delete(retiringDn);
            }
        }

        // The CASCADE constraint will clean up any remaining associations
        personRepository.delete(retiringPerson);
        LOGGER.info("Successfully merged and deleted retiring person record {}", retiringPersonId);
    }
}