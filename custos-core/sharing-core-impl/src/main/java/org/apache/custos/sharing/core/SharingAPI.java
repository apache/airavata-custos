package org.apache.custos.sharing.core;

import org.apache.custos.sharing.core.exceptions.CustosSharingException;

import java.util.List;
import java.util.Optional;

/**
 * Custos Shraing API.All sharing implementations should implement
 */
public interface SharingAPI {

    /**
     * Create an Entity Type
     *
     * @param entityType
     */
    public void createEntityType(String tenantId, EntityType entityType) throws CustosSharingException;

    /**
     * Update a given Entity Type
     *
     * @param entityType
     */
    public void updateEntityType(String tenantId, EntityType entityType) throws CustosSharingException;

    ;

    /**
     * Delete a given Entity
     *
     * @param tenantId
     * @param entityTypeId
     */
    public void deleteEntityType(String tenantId, String entityTypeId) throws CustosSharingException;

    ;

    /**
     * Get given Entity Type
     *
     * @param tenantId
     * @param entityTypeId
     * @return
     */
    public Optional<EntityType> getEntityType(String tenantId, String entityTypeId) throws CustosSharingException;

    ;

    /**
     * Get Entity Types
     *
     * @param tenantId
     * @return
     */
    public List<EntityType> getEntityTypes(String tenantId) throws CustosSharingException;

    ;


    /**
     * Create a given Permission Type
     *
     * @param permissionType
     */
    public void createPermissionType(PermissionType permissionType, String tenantId) throws CustosSharingException;

    ;

    /**
     * Update a given Permission Type
     *
     * @param permissionType
     */
    public void updatePermissionType(PermissionType permissionType, String tenantId) throws CustosSharingException;

    ;

    /**
     * Delete a given Permission Type
     *
     * @param tenantId
     * @param permissionTypeId
     */
    public void deletePermissionType(String tenantId, String permissionTypeId) throws CustosSharingException;

    ;

    /**
     * Get a permission Type
     *
     * @param tenantId
     * @param permissionTypeId
     * @return
     */
    public Optional<PermissionType> getPermissionType(String tenantId, String permissionTypeId) throws CustosSharingException;

    ;

    /**
     * Get all permission Types
     *
     * @param tenantId
     * @return
     */
    public List<PermissionType> getPermissionTypes(String tenantId) throws CustosSharingException;

    ;


    /**
     * Create an Entity
     *
     * @param entity
     */
    public void createEntity(Entity entity, String tenantId) throws CustosSharingException;

    ;

    /**
     * Update an given Entity
     *
     * @param entity
     */
    public void updateEntity(Entity entity, String tenantId) throws CustosSharingException;

    ;

    /**
     * Delete an given Entity
     *
     * @param tenantId
     * @param entityId
     */
    public void deleteEntity(String tenantId, String entityId) throws CustosSharingException;

    ;

    /**
     * Returns true is Entity is exist
     *
     * @param tenantId
     * @param entityId
     * @return
     */
    public boolean isEntityExists(String tenantId, String entityId) throws CustosSharingException;

    ;

    /**
     * Get Entity wit given ID
     *
     * @param tenantId
     * @param entityId
     * @return
     */
    public Optional<Entity> getEntity(String tenantId, String entityId) throws CustosSharingException;

    ;

    /**
     * Search entities according to search criteria
     *
     * @param tenantId
     * @param searchCriteriaList
     * @param associatingIdList
     * @param limit
     * @param offset
     * @param searchPermBottomUp
     * @return
     * @throws CustosSharingException
     */
    public List<Entity> searchEntities(String tenantId, List<SearchCriteria> searchCriteriaList, List<String> associatingIdList,
                                       int limit, int offset,
                                       boolean searchPermBottomUp)
            throws CustosSharingException;

    ;

    /**
     * Get list of shared users for given permission for given entityId
     *
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<String> getListOfSharedUsers(String tenantId, String entityId, String permissionTypeId)
            throws CustosSharingException;

    ;

    /**
     * Get list of directly (DIRECT CASCADING, DIRECT NON CASCADING, INDIRECT CASCADING) shared users of given entity id for given permission type id.
     *
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<String>  getListOfDirectlySharedUsers(String tenantId, String entityId, String permissionTypeId)
            throws CustosSharingException;

    ;


    /**
     * Get list of  directly (DIRECT CASCADING, DIRECT NON CASCADING)shared groups  of given entity id for given permission type id.
     *
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<String>  getListOfSharedGroups(String tenantId, String entityId, String permissionTypeId)
            throws CustosSharingException;

    ;

    /**
     * Get list of directly shared groups
     *
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<String>  getListOfDirectlySharedGroups(String tenantId, String entityId, String permissionTypeId)
            throws CustosSharingException;

    ;

    /**
     * Revoke permissions of user for given Entity
     *
     * @param tenantId
     * @param entityId
     * @param permissionType
     * @param usersList
     * @return
     */
    public boolean revokePermission(String tenantId, String entityId, String permissionType, List<String> usersList)
            throws CustosSharingException;

    ;

    /**
     * Share Entity with given user or group
     * @param tenantId
     * @param entityId
     * @param permissionType
     * @param associatingIds
     * @param cascade
     * @param ownerType
     * @param sharedBy
     * @throws CustosSharingException
     */
    public void shareEntity(String tenantId, String entityId, String permissionType,
                            List<String> associatingIds,boolean cascade,String ownerType, String sharedBy) throws CustosSharingException;




    public List<SharingMetadata> getAllDirectSharings(String tenantId) throws CustosSharingException;




    public List<SharingMetadata> getAllSharings(String tenantId, String entityId) throws CustosSharingException;



    /**
     * Check user has access for given entityId for given username
     *
     * @param tenantId
     * @param permission
     * @param username
     * @return
     */
    public boolean userHasAccess(String tenantId, String entityId, String permission, String username)
            throws CustosSharingException;

    ;


}
