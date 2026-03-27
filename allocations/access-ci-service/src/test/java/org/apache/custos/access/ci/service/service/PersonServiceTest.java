package org.apache.custos.access.ci.service.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.custos.access.ci.service.model.ClusterAccountEntity;
import org.apache.custos.access.ci.service.model.PersonDnsEntity;
import org.apache.custos.access.ci.service.model.PersonEntity;
import org.apache.custos.access.ci.service.repo.ClusterAccountRepository;
import org.apache.custos.access.ci.service.repo.PersonDnsRepository;
import org.apache.custos.access.ci.service.repo.PersonRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class PersonServiceTest {

    private final ObjectMapper objectMapper = new ObjectMapper();
    @Mock
    private PersonRepository personRepository;
    @Mock
    private PersonDnsRepository personDnsRepository;
    @Mock
    private ClusterAccountRepository clusterAccountRepository;
    private PersonService personService;

    @BeforeEach
    void setUp() {
        personService = new PersonService(personRepository, personDnsRepository, clusterAccountRepository);
    }

    @Test
    void findOrCreatePersonFromPacket_shouldReturnExistingPerson() {
        JsonNode packetBody = createValidPacketBody();
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findByAccessGlobalId("USER-GLOBAL-123")).thenReturn(Optional.of(existingPerson));

        PersonEntity result = personService.findOrCreatePersonFromPacket(packetBody);

        assertThat(result).isEqualTo(existingPerson);
        verify(personRepository).findByAccessGlobalId("USER-GLOBAL-123");
        verify(personRepository, never()).save(any(PersonEntity.class));
        verify(personDnsRepository, never()).save(any(PersonDnsEntity.class));
    }

    @Test
    void findOrCreatePersonFromPacket_shouldCreateNewPerson() {
        JsonNode packetBody = createValidPacketBody();
        when(personRepository.findByAccessGlobalId("USER-GLOBAL-123")).thenReturn(Optional.empty());
        when(personRepository.save(any(PersonEntity.class))).thenAnswer(invocation -> {
            PersonEntity person = invocation.getArgument(0);
            person.setId("person-123");
            return person;
        });

        PersonEntity result = personService.findOrCreatePersonFromPacket(packetBody);

        assertThat(result).isNotNull();
        assertThat(result.getAccessGlobalId()).isEqualTo("USER-GLOBAL-123");
        assertThat(result.getFirstName()).isEqualTo("John");
        assertThat(result.getLastName()).isEqualTo("Doe");
        assertThat(result.getEmail()).isEqualTo("john.doe@example.com");
        assertThat(result.getOrganization()).isEqualTo("Test Org");
        assertThat(result.getOrgCode()).isEqualTo("TEST");
        assertThat(result.getNsfStatusCode()).isEqualTo("ACTIVE");

        verify(personRepository).save(any(PersonEntity.class));
        verify(personDnsRepository, times(2)).save(any(PersonDnsEntity.class));
    }

    @Test
    void findOrCreatePersonFromPacket_shouldCreatePersonWithDnList() {
        JsonNode packetBody = createPacketBodyWithDnList();
        when(personRepository.findByAccessGlobalId("USER-GLOBAL-123")).thenReturn(Optional.empty());
        when(personRepository.save(any(PersonEntity.class))).thenAnswer(invocation -> {
            PersonEntity person = invocation.getArgument(0);
            person.setId("person-123");
            return person;
        });

        PersonEntity result = personService.findOrCreatePersonFromPacket(packetBody);

        assertThat(result).isNotNull();
        verify(personDnsRepository, times(2)).save(any(PersonDnsEntity.class));

        ArgumentCaptor<PersonDnsEntity> dnCaptor = ArgumentCaptor.forClass(PersonDnsEntity.class);
        verify(personDnsRepository, times(2)).save(dnCaptor.capture());
        List<PersonDnsEntity> savedDns = dnCaptor.getAllValues();
        assertThat(savedDns).hasSize(2);
        assertThat(savedDns.get(0).getDn()).isEqualTo("CN=John Doe,OU=Users,DC=example,DC=com");
        assertThat(savedDns.get(1).getDn()).isEqualTo("CN=John Doe,OU=Staff,DC=example,DC=com");
    }

    @Test
    void findOrCreatePersonFromPacket_shouldThrowExceptionForMissingUserGlobalId() {
        ObjectNode packetBody = objectMapper.createObjectNode()
                .put("UserFirstName", "John")
                .put("UserLastName", "Doe");

        assertThatThrownBy(() -> personService.findOrCreatePersonFromPacket(packetBody))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Packet body must contain a 'UserGlobalID'");
    }

    // -------------------------------------------------------------------------
    // replaceFromModifyPacket
    // -------------------------------------------------------------------------

    @Test
    void replaceFromModifyPacket_shouldUpdatePersonFields() {
        JsonNode body = createModifyPacketBody();
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        verify(personRepository).save(existingPerson);
        assertThat(existingPerson.getFirstName()).isEqualTo("Jane");
        assertThat(existingPerson.getLastName()).isEqualTo("Smith");
        assertThat(existingPerson.getEmail()).isEqualTo("jane.smith@example.com");
        assertThat(existingPerson.getOrganization()).isEqualTo("New Org");
        assertThat(existingPerson.getOrgCode()).isEqualTo("NEW");
        assertThat(existingPerson.getNsfStatusCode()).isEqualTo("INACTIVE");
    }

    @Test
    void replaceFromModifyPacket_shouldUseCorrectAmieFieldNames() {
        // Verify that the correct AMIE field names are used (UserFirstName not FirstName, etc.)
        ObjectNode body = objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "Updated")
                .put("UserLastName", "Name")
                .put("UserEmail", "updated@example.com")
                .put("UserOrganization", "Updated Org")
                .put("UserOrgCode", "UPD")
                .put("NsfStatusCode", "INACTIVE");

        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        assertThat(existingPerson.getFirstName()).isEqualTo("Updated");
        assertThat(existingPerson.getLastName()).isEqualTo("Name");
        assertThat(existingPerson.getEmail()).isEqualTo("updated@example.com");
        assertThat(existingPerson.getOrganization()).isEqualTo("Updated Org");
        assertThat(existingPerson.getOrgCode()).isEqualTo("UPD");
        assertThat(existingPerson.getNsfStatusCode()).isEqualTo("INACTIVE");
    }

    @Test
    void replaceFromModifyPacket_shouldNotWipeFieldsAbsentFromPacket() {
        // Packet only contains PersonID and UserFirstName — other fields must remain unchanged
        ObjectNode body = objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "OnlyFirstUpdated");

        PersonEntity existingPerson = createPersonEntity();
        // Pre-existing values that must survive
        existingPerson.setLastName("OriginalLast");
        existingPerson.setEmail("original@example.com");
        existingPerson.setOrganization("Original Org");
        existingPerson.setOrgCode("ORIG");
        existingPerson.setNsfStatusCode("ACTIVE");

        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        assertThat(existingPerson.getFirstName()).isEqualTo("OnlyFirstUpdated");
        assertThat(existingPerson.getLastName()).isEqualTo("OriginalLast");
        assertThat(existingPerson.getEmail()).isEqualTo("original@example.com");
        assertThat(existingPerson.getOrganization()).isEqualTo("Original Org");
        assertThat(existingPerson.getOrgCode()).isEqualTo("ORIG");
        assertThat(existingPerson.getNsfStatusCode()).isEqualTo("ACTIVE");
    }

    @Test
    void replaceFromModifyPacket_shouldNotWipeOrganizationWhenAbsent() {
        // Regression test: old code unconditionally wiped organization/orgCode to null when field absent
        ObjectNode body = objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "Jane");
        // No UserOrganization, UserOrgCode in packet

        PersonEntity existingPerson = createPersonEntity();
        existingPerson.setOrganization("Keep This Org");
        existingPerson.setOrgCode("KEEP");

        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        assertThat(existingPerson.getOrganization()).isEqualTo("Keep This Org");
        assertThat(existingPerson.getOrgCode()).isEqualTo("KEEP");
    }

    @Test
    void replaceFromModifyPacket_shouldNotClearDnsWhenUserDnListAbsent() {
        // When UserDnList is not present in the packet, existing DNs must not be touched
        ObjectNode body = objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "Jane");
        // No UserDnList field

        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        // deleteByPerson_Id should NOT have been called since UserDnList was absent
        verify(personDnsRepository, never()).deleteByPerson_Id("person-123");
    }

    @Test
    void replaceFromModifyPacket_shouldUpdateDnList() {
        JsonNode body = createModifyPacketBodyWithDns();
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));
        when(personDnsRepository.existsByPerson_IdAndDn("person-123", "CN=Jane Smith,OU=Users,DC=example,DC=com")).thenReturn(false);

        personService.replaceFromModifyPacket(body);

        //noinspection unchecked
        verify(personDnsRepository).deleteByPerson_IdAndDnNotIn(eq("person-123"), any(List.class));
        verify(personDnsRepository).save(any(PersonDnsEntity.class));
    }

    @Test
    void replaceFromModifyPacket_shouldClearDnsWhenEmptyUserDnListPresent() {
        // When UserDnList is present but empty, existing DNs should be cleared
        ObjectNode body = objectMapper.createObjectNode()
                .put("PersonID", "person-123");
        body.set("UserDnList", objectMapper.createArrayNode()); // explicitly empty array

        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));

        personService.replaceFromModifyPacket(body);

        verify(personDnsRepository).deleteByPerson_Id("person-123");
    }

    @Test
    void replaceFromModifyPacket_shouldThrowExceptionForMissingPersonId() {
        ObjectNode body = objectMapper.createObjectNode().put("UserFirstName", "Jane");

        assertThatThrownBy(() -> personService.replaceFromModifyPacket(body))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Missing required 'PersonID'");
    }

    @Test
    void replaceFromModifyPacket_shouldThrowExceptionForUnknownPersonId() {
        JsonNode body = createModifyPacketBody();
        when(personRepository.findById("person-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> personService.replaceFromModifyPacket(body))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Unknown local PersonID: person-123");
    }

    // -------------------------------------------------------------------------
    // persistDnsForPerson tests
    // -------------------------------------------------------------------------

    @Test
    void persistDnsForPerson_shouldPersistNewDns() {
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));
        when(personDnsRepository.existsByPerson_IdAndDn("person-123", "/C=US/O=Test/CN=John Doe")).thenReturn(false);

        JsonNode dnList = objectMapper.createArrayNode().add("/C=US/O=Test/CN=John Doe");
        personService.persistDnsForPerson("person-123", dnList);

        ArgumentCaptor<PersonDnsEntity> captor = ArgumentCaptor.forClass(PersonDnsEntity.class);
        verify(personDnsRepository).save(captor.capture());
        assertThat(captor.getValue().getDn()).isEqualTo("/C=US/O=Test/CN=John Doe");
    }

    @Test
    void persistDnsForPerson_shouldBeIdempotentWhenDnAlreadyExists() {
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));
        when(personDnsRepository.existsByPerson_IdAndDn("person-123", "/C=US/O=Test/CN=John Doe")).thenReturn(true);

        JsonNode dnList = objectMapper.createArrayNode().add("/C=US/O=Test/CN=John Doe");
        personService.persistDnsForPerson("person-123", dnList);

        verify(personDnsRepository, never()).save(any(PersonDnsEntity.class));
    }

    @Test
    void persistDnsForPerson_shouldPersistOnlyNewDnsWhenSomeDnsAlreadyExist() {
        PersonEntity existingPerson = createPersonEntity();
        when(personRepository.findById("person-123")).thenReturn(Optional.of(existingPerson));
        when(personDnsRepository.existsByPerson_IdAndDn("person-123", "/C=US/O=Existing/CN=John")).thenReturn(true);
        when(personDnsRepository.existsByPerson_IdAndDn("person-123", "/C=US/O=New/CN=John")).thenReturn(false);

        JsonNode dnList = objectMapper.createArrayNode()
                .add("/C=US/O=Existing/CN=John")
                .add("/C=US/O=New/CN=John");
        personService.persistDnsForPerson("person-123", dnList);

        ArgumentCaptor<PersonDnsEntity> captor = ArgumentCaptor.forClass(PersonDnsEntity.class);
        verify(personDnsRepository, times(1)).save(captor.capture());
        assertThat(captor.getValue().getDn()).isEqualTo("/C=US/O=New/CN=John");
    }

    @Test
    void persistDnsForPerson_shouldDoNothingForNullDnList() {
        personService.persistDnsForPerson("person-123", null);

        verify(personRepository, never()).findById(any());
        verify(personDnsRepository, never()).save(any());
    }

    @Test
    void persistDnsForPerson_shouldDoNothingForEmptyDnList() {
        JsonNode emptyDnList = objectMapper.createArrayNode();
        personService.persistDnsForPerson("person-123", emptyDnList);

        verify(personRepository, never()).findById(any());
        verify(personDnsRepository, never()).save(any());
    }

    @Test
    void persistDnsForPerson_shouldThrowForUnknownPersonId() {
        when(personRepository.findById("unknown-id")).thenReturn(Optional.empty());
        JsonNode dnList = objectMapper.createArrayNode().add("/C=US/O=Test/CN=John");

        assertThatThrownBy(() -> personService.persistDnsForPerson("unknown-id", dnList))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Unknown local PersonID: unknown-id");
    }

    // -------------------------------------------------------------------------
    // deleteFromModifyPacket tests
    // -------------------------------------------------------------------------

    @Test
    void deleteFromModifyPacket_shouldDeletePerson() {
        JsonNode body = createModifyPacketBody();
        personService.deleteFromModifyPacket(body);
        verify(personRepository).deleteById("person-123");
    }

    @Test
    void deleteFromModifyPacket_shouldThrowExceptionForMissingPersonId() {
        ObjectNode body = objectMapper.createObjectNode().put("UserFirstName", "Jane");

        assertThatThrownBy(() -> personService.deleteFromModifyPacket(body))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Missing required 'PersonID'");
    }

    // -------------------------------------------------------------------------
    // mergePersons tests
    // -------------------------------------------------------------------------

    @Test
    void mergePersons_shouldMergeSuccessfully() {
        PersonEntity survivingPerson = createPersonEntity();
        PersonEntity retiringPerson = createPersonEntity();
        retiringPerson.setId("retiring-person-123");

        ClusterAccountEntity clusterAccount = new ClusterAccountEntity();
        clusterAccount.setId("account-123");
        clusterAccount.setUsername("jdoe");
        clusterAccount.setPerson(retiringPerson);
        retiringPerson.setClusterAccounts(List.of(clusterAccount));

        when(personRepository.findById("surviving-person-123")).thenReturn(Optional.of(survivingPerson));
        when(personRepository.findById("retiring-person-123")).thenReturn(Optional.of(retiringPerson));

        personService.mergePersons("surviving-person-123", "retiring-person-123");

        verify(clusterAccountRepository).save(clusterAccount);
        assertThat(clusterAccount.getPerson()).isEqualTo(survivingPerson);
        verify(personRepository).delete(retiringPerson);
    }

    @Test
    void mergePersons_shouldThrowExceptionForUnknownSurvivingPerson() {
        when(personRepository.findById("surviving-person-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> personService.mergePersons("surviving-person-123", "retiring-person-123"))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("Could not find surviving person with local ID: surviving-person-123");
    }

    @Test
    void mergePersons_shouldThrowExceptionForUnknownRetiringPerson() {
        PersonEntity survivingPerson = createPersonEntity();
        when(personRepository.findById("surviving-person-123")).thenReturn(Optional.of(survivingPerson));
        when(personRepository.findById("retiring-person-123")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> personService.mergePersons("surviving-person-123", "retiring-person-123"))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("Could not find retiring person with local ID: retiring-person-123");
    }

    // -------------------------------------------------------------------------
    // Test data helpers
    // -------------------------------------------------------------------------

    private JsonNode createValidPacketBody() {
        return objectMapper.createObjectNode()
                .put("UserGlobalID", "USER-GLOBAL-123")
                .put("UserFirstName", "John")
                .put("UserLastName", "Doe")
                .put("UserEmail", "john.doe@example.com")
                .put("UserOrganization", "Test Org")
                .put("UserOrgCode", "TEST")
                .put("NsfStatusCode", "ACTIVE")
                .set("UserDnList", objectMapper.createArrayNode()
                        .add("CN=John Doe,OU=Users,DC=example,DC=com")
                        .add("CN=John Doe,OU=Staff,DC=example,DC=com"));
    }

    private JsonNode createPacketBodyWithDnList() {
        return createValidPacketBody();
    }

    private JsonNode createModifyPacketBody() {
        return objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "Jane")
                .put("UserLastName", "Smith")
                .put("UserEmail", "jane.smith@example.com")
                .put("UserOrganization", "New Org")
                .put("UserOrgCode", "NEW")
                .put("NsfStatusCode", "INACTIVE");
    }

    private JsonNode createModifyPacketBodyWithDns() {
        return objectMapper.createObjectNode()
                .put("PersonID", "person-123")
                .put("UserFirstName", "Jane")
                .put("UserLastName", "Smith")
                .set("UserDnList", objectMapper.createArrayNode().add("CN=Jane Smith,OU=Users,DC=example,DC=com"));
    }

    private PersonEntity createPersonEntity() {
        PersonEntity entity = new PersonEntity();
        entity.setId("person-123");
        entity.setAccessGlobalId("USER-GLOBAL-123");
        entity.setFirstName("John");
        entity.setLastName("Doe");
        entity.setEmail("john.doe@example.com");
        entity.setOrganization("Test Org");
        entity.setOrgCode("TEST");
        entity.setNsfStatusCode("ACTIVE");
        return entity;
    }
}
