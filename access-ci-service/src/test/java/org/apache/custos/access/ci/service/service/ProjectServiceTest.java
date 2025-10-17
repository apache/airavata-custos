package org.apache.custos.access.ci.service.service;

import org.apache.custos.access.ci.service.model.ProjectEntity;
import org.apache.custos.access.ci.service.repo.ProjectRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class ProjectServiceTest {

    @Mock
    private ProjectRepository projectRepository;

    private ProjectService projectService;

    @BeforeEach
    void setUp() {
        projectService = new ProjectService(projectRepository);
    }

    @Test
    void createOrFindProject_shouldReturnExistingProject() {
        ProjectEntity existingProject = createProjectEntity();
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(existingProject));

        ProjectEntity result = projectService.createOrFindProject("PROJECT-123", "GRANT-456");

        assertThat(result).isEqualTo(existingProject);
        verify(projectRepository).findById("PROJECT-123");
        verify(projectRepository, never()).save(any(ProjectEntity.class));
    }

    @Test
    void createOrFindProject_shouldCreateNewProject() {
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.empty());
        when(projectRepository.save(any(ProjectEntity.class))).thenAnswer(invocation -> invocation.<ProjectEntity>getArgument(0));

        ProjectEntity result = projectService.createOrFindProject("PROJECT-123", "GRANT-456");

        assertThat(result).isNotNull();
        assertThat(result.getId()).isEqualTo("PROJECT-123");
        assertThat(result.getGrantNumber()).isEqualTo("GRANT-456");
        assertThat(result.isActive()).isTrue();

        verify(projectRepository).save(any(ProjectEntity.class));
    }

    @Test
    void inactivateProject_shouldInactivateExistingProject() {
        ProjectEntity project = createProjectEntity();
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(project));

        projectService.inactivateProject("PROJECT-123");

        assertThat(project.isActive()).isFalse();
        verify(projectRepository).save(project);
    }

    @Test
    void inactivateProject_shouldThrowExceptionForNonExistentProject() {
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> projectService.inactivateProject("PROJECT-123"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Project not found: PROJECT-123");
    }

    @Test
    void reactivateProject_shouldReactivateExistingProject() {
        ProjectEntity project = createProjectEntity();
        project.setActive(false);
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.of(project));

        projectService.reactivateProject("PROJECT-123");

        assertThat(project.isActive()).isTrue();
        verify(projectRepository).save(project);
    }

    @Test
    void reactivateProject_shouldThrowExceptionForNonExistentProject() {
        when(projectRepository.findById("PROJECT-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> projectService.reactivateProject("PROJECT-123"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Project not found: PROJECT-123");
    }

    private ProjectEntity createProjectEntity() {
        ProjectEntity entity = new ProjectEntity();
        entity.setId("PROJECT-123");
        entity.setGrantNumber("GRANT-456");
        entity.setActive(true);
        return entity;
    }
}
