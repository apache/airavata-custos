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

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
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
    void provisionClusterAccount_shouldCreateUniqueUsername() {
        PersonEntity person = createPersonEntity();
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

        assertThatThrownBy(() -> userAccountService.provisionClusterAccount(person)).isInstanceOf(StringIndexOutOfBoundsException.class);
    }

    private PersonEntity createPersonEntity() {
        PersonEntity entity = new PersonEntity();
        entity.setId("person-123");
        entity.setAccessGlobalId("12345");
        entity.setFirstName("John");
        entity.setLastName("Doe");
        entity.setEmail("john.doe@example.com");
        return entity;
    }
}
