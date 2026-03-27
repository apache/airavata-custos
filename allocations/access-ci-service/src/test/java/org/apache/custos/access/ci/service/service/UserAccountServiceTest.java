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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class UserAccountServiceTest {

    @Mock
    private ClusterAccountRepository clusterAccountRepository;

    private UserAccountService userAccountService;

    @BeforeEach
    void setUp() {
        userAccountService = new UserAccountService(clusterAccountRepository);
    }

    @Test
    void provisionClusterAccount_shouldReturnExistingAccountWhenPersonAlreadyHasOne() {
        PersonEntity person = createPersonEntity();
        ClusterAccountEntity existingAccount = createClusterAccountEntity(person, "jdoe");
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of(existingAccount));

        ClusterAccountEntity result = userAccountService.provisionClusterAccount(person);

        assertThat(result).isSameAs(existingAccount);
        assertThat(result.getUsername()).isEqualTo("jdoe");
        verify(clusterAccountRepository, never()).save(any(ClusterAccountEntity.class));
    }

    @Test
    void provisionClusterAccount_calledTwiceForSamePerson_shouldReturnSameAccountBothTimes() {
        PersonEntity person = createPersonEntity();
        ClusterAccountEntity existingAccount = createClusterAccountEntity(person, "jdoe");

        // First call: no account exists yet, create one.
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());
        when(clusterAccountRepository.findByUsername("jdoe")).thenReturn(Optional.empty());
        when(clusterAccountRepository.save(any(ClusterAccountEntity.class)))
                .thenAnswer(invocation -> invocation.<ClusterAccountEntity>getArgument(0));

        ClusterAccountEntity firstResult = userAccountService.provisionClusterAccount(person);
        assertThat(firstResult.getUsername()).isEqualTo("jdoe");

        // Second call: account now exists.
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of(existingAccount));

        ClusterAccountEntity secondResult = userAccountService.provisionClusterAccount(person);
        assertThat(secondResult).isSameAs(existingAccount);
        assertThat(secondResult.getUsername()).isEqualTo("jdoe");
    }

    // -------------------------------------------------------------------------
    // New account creation tests
    // -------------------------------------------------------------------------

    @Test
    void provisionClusterAccount_shouldCreateUniqueUsername() {
        PersonEntity person = createPersonEntity();
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());
        when(clusterAccountRepository.findByUsername("jdoe")).thenReturn(Optional.empty());
        when(clusterAccountRepository.save(any(ClusterAccountEntity.class)))
                .thenAnswer(invocation -> invocation.<ClusterAccountEntity>getArgument(0));

        ClusterAccountEntity result = userAccountService.provisionClusterAccount(person);

        assertThat(result).isNotNull();
        assertThat(result.getUsername()).isEqualTo("jdoe");
        assertThat(result.getPerson()).isEqualTo(person);
        verify(clusterAccountRepository).save(result);
    }

    @Test
    void provisionClusterAccount_shouldGenerateUniqueUsernameWhenBaseIsTaken() {
        PersonEntity person = createPersonEntity();
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());
        when(clusterAccountRepository.findByUsername("jdoe")).thenReturn(Optional.of(new ClusterAccountEntity()));
        when(clusterAccountRepository.findByUsername("jdoe1")).thenReturn(Optional.empty());
        when(clusterAccountRepository.save(any(ClusterAccountEntity.class)))
                .thenAnswer(invocation -> invocation.<ClusterAccountEntity>getArgument(0));

        ClusterAccountEntity result = userAccountService.provisionClusterAccount(person);

        assertThat(result).isNotNull();
        assertThat(result.getUsername()).isEqualTo("jdoe1");
        assertThat(result.getPerson()).isEqualTo(person);
        verify(clusterAccountRepository).save(result);
    }

    @Test
    void provisionClusterAccount_shouldHandleMultipleSuffixes() {
        PersonEntity person = createPersonEntity();
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());
        when(clusterAccountRepository.findByUsername("jdoe")).thenReturn(Optional.of(new ClusterAccountEntity()));
        when(clusterAccountRepository.findByUsername("jdoe1")).thenReturn(Optional.of(new ClusterAccountEntity()));
        when(clusterAccountRepository.findByUsername("jdoe2")).thenReturn(Optional.empty());
        when(clusterAccountRepository.save(any(ClusterAccountEntity.class)))
                .thenAnswer(invocation -> invocation.<ClusterAccountEntity>getArgument(0));

        ClusterAccountEntity result = userAccountService.provisionClusterAccount(person);

        assertThat(result).isNotNull();
        assertThat(result.getUsername()).isEqualTo("jdoe2");
        assertThat(result.getPerson()).isEqualTo(person);
        verify(clusterAccountRepository).save(result);
    }

    @Test
    void provisionClusterAccount_shouldHandleNamesWithSpaces() {
        PersonEntity person = createPersonEntity();
        person.setFirstName("John Michael");
        person.setLastName("Doe Smith");
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());
        when(clusterAccountRepository.findByUsername("jdoe-smith")).thenReturn(Optional.empty());
        when(clusterAccountRepository.save(any(ClusterAccountEntity.class)))
                .thenAnswer(invocation -> invocation.<ClusterAccountEntity>getArgument(0));

        ClusterAccountEntity result = userAccountService.provisionClusterAccount(person);

        assertThat(result).isNotNull();
        assertThat(result.getUsername()).isEqualTo("jdoe-smith");
        assertThat(result.getPerson()).isEqualTo(person);
        verify(clusterAccountRepository).save(result);
    }

    @Test
    void provisionClusterAccount_shouldThrowExceptionForEmptyNames() {
        PersonEntity person = createPersonEntity();
        person.setFirstName("");
        person.setLastName("");
        when(clusterAccountRepository.findByPerson(person)).thenReturn(List.of());

        assertThatThrownBy(() -> userAccountService.provisionClusterAccount(person)).isInstanceOf(StringIndexOutOfBoundsException.class);
    }

    // -------------------------------------------------------------------------
    // Helper factories
    // -------------------------------------------------------------------------

    private PersonEntity createPersonEntity() {
        PersonEntity entity = new PersonEntity();
        entity.setId("person-123");
        entity.setAccessGlobalId("12345");
        entity.setFirstName("John");
        entity.setLastName("Doe");
        entity.setEmail("john.doe@example.com");
        return entity;
    }

    private ClusterAccountEntity createClusterAccountEntity(PersonEntity person, String username) {
        ClusterAccountEntity entity = new ClusterAccountEntity();
        entity.setId("account-456");
        entity.setPerson(person);
        entity.setUsername(username);
        return entity;
    }
}
