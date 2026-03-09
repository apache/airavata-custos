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
    void createMembership_shouldReturnExistingMembership() {
        ProjectMembershipEntity existingMembership = createMembershipEntity();
        when(membershipRepository.findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456"))
                .thenReturn(Optional.of(existingMembership));

        ProjectMembershipEntity result = membershipService.createMembership("PROJECT-123", "ACCOUNT-456", "PI");

        assertThat(result).isEqualTo(existingMembership);
        verify(membershipRepository).findByProjectIdAndClusterAccountId("PROJECT-123", "ACCOUNT-456");
        verify(membershipRepository, never()).save(any(ProjectMembershipEntity.class));
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

    @Test
    void inactivateMembershipsByPersonAndProject_shouldInactivateMemberships() {
        ProjectMembershipEntity membership1 = createMembershipEntity();
        ProjectMembershipEntity membership2 = createMembershipEntity();
        membership2.setId("membership-2");

        List<ProjectMembershipEntity> memberships = List.of(membership1, membership2);
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(memberships);

        membershipService.inactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");

        assertThat(membership1.isActive()).isFalse();
        assertThat(membership2.isActive()).isFalse();
        verify(membershipRepository).saveAll(memberships);
    }

    @Test
    void inactivateMembershipsByPersonAndProject_shouldHandleNoMemberships() {
        when(membershipRepository.findByProjectIdAndClusterAccount_Person_Id("PROJECT-123", "PERSON-456")).thenReturn(List.of());
        membershipService.inactivateMembershipsByPersonAndProject("PROJECT-123", "PERSON-456");
        verify(membershipRepository, never()).saveAll(any());
    }

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
