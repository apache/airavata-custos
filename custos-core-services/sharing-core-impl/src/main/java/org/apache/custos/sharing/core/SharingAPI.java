package org.apache.custos.sharing.core;

import java.util.List;

/**
 * Custos Shraing API.All sharing implementations should implement
 */
public interface SharingAPI {

    /**
     * Create an Entity Type
     * @param entityType
     */
    public void createEntityType(String tenantId, EntityType entityType);

    /**
     * Update a given Entity Type
     * @param entityType
     */
    public  void updateEntityType(String tenantId,EntityType entityType);

    /**
     * Delete a given Entity
     * @param tenantId
     * @param entityTypeId
     */
    public  void deleteEntityType(String tenantId, String entityTypeId);

    /**
     * Get given Entity Type
     * @param tenantId
     * @param entityTypeId
     * @return
     */
    public EntityType getEntityType(String tenantId, String entityTypeId);

    /**
     * Get Entity Types
     * @param tenantId
     * @return
     */
    public List<EntityType> getEntityTypes(String tenantId);


    /**
     * Create a given Permission Type
     * @param permissionType
     */
    public void createPermissionType(PermissionType permissionType);

    /**
     * Update a given Permission Type
     * @param permissionType
     */
    public  void updatePermissionType(PermissionType permissionType);

    /**
     * Delete a given Permission Type
     * @param tenantId
     * @param permissionTypeId
     */
    public  void deletePermissionType(String tenantId, String permissionTypeId);

    /**
     * Get a permission Type
     * @param tenantId
     * @param permissionTypeId
     * @return
     */
    public EntityType getPermissionType(String tenantId, String permissionTypeId);

    /**
     * Get all permission Types
     * @param tenantId
     * @return
     */
    public List<EntityType> getPermissionTypes(String tenantId);


    /**
     * Create an Entity
     * @param entity
     */
    public void createEntity(Entity entity);

    /**
     * Update an given Entity
     * @param entity
     */
    public  void updateEntity(Entity entity);

    /**
     * Delete an given Entity
     * @param tenantId
     * @param entityId
     */
    public  void deleteEntity(String tenantId, String entityId);

    /**
     * Returns true is Entity is exist
     * @param tenantId
     * @param entityId
     * @return
     */
    public boolean isEntityExists(String tenantId, String entityId);

    /**
     * Get Entity wit given ID
     * @param tenantId
     * @param entityId
     * @return
     */
    public Entity getEntity(String tenantId, String entityId);

    /**
     * Search Entities for given search criteria
     * @param tenantId
     * @param searchCriteriaList
     * @param limit
     * @param offset
     * @return
     */
    public List<Entity> searchEntities(String tenantId, List<SearchCriteria> searchCriteriaList,int limit, int offset);

    /**
     * Get list of shared users for given permission for given entityId
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<SharedOwners> getListOfSharedUsers(String tenantId, String entityId, String permissionTypeId);

    /**
     * Get list of directly (DIRECT CASCADING, DIRECT NON CASCADING, INDIRECT CASCADING) shared users of given entity id for given permission type id.
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<SharedOwners> getListOfDirectlySharedUsers(String tenantId, String entityId, String permissionTypeId);


    /**
     * Get list of  directly (DIRECT CASCADING, DIRECT NON CASCADING)shared groups  of given entity id for given permission type id.
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<SharedOwners> getListOfSharedGroups(String tenantId, String entityId, String permissionTypeId);

    /**
     * Get list of directly shared groups
     * @param tenantId
     * @param entityId
     * @param permissionTypeId
     * @return
     */
    public List<SharedOwners> getListOfDirectlySharedGroups(String tenantId, String entityId, String permissionTypeId);

    /**
     * Revoke permissions of user for given Entity
     * @param tenantId
     * @param entityId
     * @param permissionType
     * @param usersList
     * @return
     */
    public boolean revokePermission(String tenantId, String entityId, String permissionType, List<String> usersList);


    /**
     * Check user has access for given entityId for given username
     * @param tenantId
     * @param permission
     * @param username
     * @return
     */
    public boolean userHasAccess(String tenantId, String entityId, String permission, String username);



}
