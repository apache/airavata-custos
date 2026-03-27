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

import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.ProjectEntity;
import org.apache.custos.access.ci.service.model.ProjectMembershipEntity;
import org.apache.custos.access.ci.service.repo.ClusterAccountRepository;
import org.apache.custos.access.ci.service.repo.ProjectMembershipRepository;
import org.apache.custos.access.ci.service.repo.ProjectRepository;
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
class ProjectMembershipServiceTest {

    @Mock
    private ProjectMembershipRepository membershipRepository;
    @Mock
    private ProjectRepository projectRepository;
    @Mock
    private ClusterAccountRepository clusterAccountRepository;

    private ProjectMembershipService membershipService;

    @BeforeEach
    void setUp() {
        membershipService = new ProjectMembershipService(membershipRepository, projectRepository, clusterAccountRepository);
    }

    @Test
    void createMembership_shouldReturnExistingActiveMembership() {
        ProjectMembershipEntity existingMembership = createMembershipEntity();
        existingMembership.setActive(true);
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456"))
                .thenReturn(Optional.of(existingMembership));

        ProjectMembershipEntity result = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI");

        assertThat(result).isEqualTo(existingMembership);
        assertThat(result.isActive()).isTrue();
        verify(membershipRepository).findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456");
        verify(membershipRepository, never()).save(any(ProjectMembershipEntity.class));
    }

    @Test
    void createMembership_shouldReactivateInactiveMembership() {
        ProjectMembershipEntity inactiveMembership = createMembershipEntity();
        inactiveMembership.setActive(false);
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456"))
                .thenReturn(Optional.of(inactiveMembership));
        when(membershipRepository.save(any(ProjectMembershipEntity.class)))
                .thenAnswer(invocation -> invocation.<ProjectMembershipEntity>getArgument(0));

        ProjectMembershipEntity result = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "USER");

        assertThat(result).isSameAs(inactiveMembership);
        assertThat(result.isActive()).isTrue();
        assertThat(result.getRole()).isEqualTo("USER");
        verify(membershipRepository).save(inactiveMembership);
    }

    @Test
    void createMembership_calledTwiceForSamePair_shouldNotCreateDuplicate() {
        ProjectEntity project = createProjectEntity();
        ClusterAccountEntity clusterAccount = createClusterAccountEntity();

        // First call: no membership exists.
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456"))
                .thenReturn(Optional.empty());
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(project));
        when(clusterAccountRepository.findById("ACCOUNT-456")).thenReturn(Optional.of(clusterAccount));
        when(membershipRepository.save(any(ProjectMembershipEntity.class)))
                .thenAnswer(invocation -> invocation.<ProjectMembershipEntity>getArgument(0));

        ProjectMembershipEntity firstResult = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI");
        assertThat(firstResult.isActive()).isTrue();

        // Second call: membership now exists and is active.
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456"))
                .thenReturn(Optional.of(firstResult));

        ProjectMembershipEntity secondResult = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI");
        assertThat(secondResult).isSameAs(firstResult);
    }

    @Test
    void createMembership_shouldCreateNewMembership() {
        ProjectEntity project = createProjectEntity();
        ClusterAccountEntity clusterAccount = createClusterAccountEntity();

        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456")).thenReturn(Optional.empty());
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(project));
        when(clusterAccountRepository.findById("ACCOUNT-456")).thenReturn(Optional.of(clusterAccount));
        when(membershipRepository.save(any(ProjectMembershipEntity.class)))
                .thenAnswer(invocation -> invocation.<ProjectMembershipEntity>getArgument(0));

        ProjectMembershipEntity result = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI");

        assertThat(result).isNotNull();
        assertThat(result.getProject()).isEqualTo(project);
        assertThat(result.getClusterAccount()).isEqualTo(clusterAccount);
        assertThat(result.getRole()).isEqualTo("PI");
        assertThat(result.isActive()).isTrue();

        verify(membershipRepository).save(result);
    }

    @Test
    void createMembership_shouldThrowExceptionForNonExistentProject() {
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456")).thenReturn(Optional.empty());
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Project not found: PROJECT-123");
    }

    @Test
    void createMembership_shouldThrowExceptionForNonExistentClusterAccount() {
        ProjectEntity project = createProjectEntity();
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456")).thenReturn(Optional.empty());
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(project));
        when(clusterAccountRepository.findById("ACCOUNT-456")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Cluster account not found: ACCOUNT-456");
    }

    // -------------------------------------------------------------------------
    // inactivateMembershipsByPersonAndProject tests
    // -------------------------------------------------------------------------

    @Test
    void inactivateMembershipsByPersonAndProject_shouldInactivateMembershipsAndReturnCount() {
        ProjectMembershipEntity membership1 = createMembershipEntity();
        ProjectMembershipEntity membership2 = createMembershipEntity();
        membership2.setId("membership-2");

        List<ProjectMembershipEntity> memberships = List.of(membership1, membership2);
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(memberships);

        int count = membershipService.inactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");

        assertThat(count).isEqualTo(2);
        assertThat(membership1.isActive()).isFalse();
        assertThat(membership2.isActive()).isFalse();
        verify(membershipRepository).saveAll(memberships);
    }

    @Test
    void inactivateMembershipsByPersonAndProject_shouldReturnZeroWhenNoMembershipsFound() {
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(List.of());

        int count = membershipService.inactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");

        assertThat(count).isZero();
        verify(membershipRepository, never()).saveAll(any());
    }

    // -------------------------------------------------------------------------
    // inactivateAllMembershipsForProject tests
    // -------------------------------------------------------------------------

    @Test
    void inactivateAllMembershipsForProject_shouldInactivateAllMemberships() {
        ProjectMembershipEntity membership1 = createMembershipEntity();
        ProjectMembershipEntity membership2 = createMembershipEntity();
        membership2.setId("membership-2");

        List<ProjectMembershipEntity> memberships = List.of(membership1, membership2);
        when(membershipRepository.findByProjectId("PROJECT-123")).thenReturn(memberships);

        membershipService.inactivateAllMembershipsForProject("PROJECT-123");

        assertThat(membership1.isActive()).isFalse();
        assertThat(membership2.isActive()).isFalse();
        verify(membershipRepository).saveAll(memberships);
    }

    @Test
    void inactivateAllMembershipsForProject_shouldHandleNoMemberships() {
        when(membershipRepository.findByProjectId("PROJECT-123")).thenReturn(List.of());
        membershipService.inactivateAllMembershipsForProject("PROJECT-123");
        verify(membershipRepository, never()).saveAll(any());
    }

    // -------------------------------------------------------------------------
    // reactivatePiMembership tests
    // -------------------------------------------------------------------------

    @Test
    void reactivatePiMembership_shouldReactivatePiMemberships() {
        ProjectMembershipEntity piMembership1 = createMembershipEntity();
        piMembership1.setRole("PI");
        piMembership1.setActive(false);

        ProjectMembershipEntity piMembership2 = createMembershipEntity();
        piMembership2.setId("membership-2");
        piMembership2.setRole("PI");
        piMembership2.setActive(false);

        List<ProjectMembershipEntity> piMemberships = List.of(piMembership1, piMembership2);
        when(membershipRepository.findByProjectIdAndRole("PROJECT-123", "PI")).thenReturn(piMemberships);

        membershipService.reactivatePiMembership("PROJECT-123");

        assertThat(piMembership1.isActive()).isTrue();
        assertThat(piMembership2.isActive()).isTrue();
        verify(membershipRepository).saveAll(piMemberships);
    }

    @Test
    void reactivatePiMembership_shouldHandleNoPiMemberships() {
        when(membershipRepository.findByProjectIdAndRole("PROJECT-123", "PI")).thenReturn(List.of());
        membershipService.reactivatePiMembership("PROJECT-123");
        verify(membershipRepository).saveAll(List.of());
    }

    // -------------------------------------------------------------------------
    // reactivateMembershipsByPersonAndProject tests
    // -------------------------------------------------------------------------

    @Test
    void reactivateMembershipsByPersonAndProject_shouldReactivateMembershipsAndReturnCount() {
        ProjectMembershipEntity membership1 = createMembershipEntity();
        membership1.setActive(false);
        ProjectMembershipEntity membership2 = createMembershipEntity();
        membership2.setId("membership-2");
        membership2.setActive(false);

        List<ProjectMembershipEntity> memberships = List.of(membership1, membership2);
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(memberships);

        int count = membershipService.reactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");

        assertThat(count).isEqualTo(2);
        assertThat(membership1.isActive()).isTrue();
        assertThat(membership2.isActive()).isTrue();
        verify(membershipRepository).saveAll(memberships);
    }

    @Test
    void reactivateMembershipsByPersonAndProject_shouldReturnZeroWhenNoMembershipsFound() {
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(List.of());

        int count = membershipService.reactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");

        assertThat(count).isZero();
        verify(membershipRepository, never()).saveAll(any());
    }

    private ProjectEntity createProjectEntity() {
        ProjectEntity entity = new ProjectEntity();
        entity.setId("PROJECT-123");
        entity.setGrantNumber("GRANT-456");
        entity.setActive(true);
        return entity;
    }

    private ClusterAccountEntity createClusterAccountEntity() {
        ClusterAccountEntity entity = new ClusterAccountEntity();
        entity.setId("ACCOUNT-456");
        entity.setUsername("jdoe");
        return entity;
    }

    private ProjectMembershipEntity createMembershipEntity() {
        ProjectMembershipEntity entity = new ProjectMembershipEntity();
        entity.setId("membership-123");
        entity.setProject(createProjectEntity());
        entity.setClusterAccount(createClusterAccountEntity());
        entity.setRole("PI");
        entity.setActive(true);
        return entity;
    }
}
