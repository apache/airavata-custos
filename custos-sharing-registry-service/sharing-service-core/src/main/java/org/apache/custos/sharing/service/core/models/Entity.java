package org.apache.custos.sharing.service.core.models;

import javax.validation.constraints.NotNull;

public class Entity {
    
    @NotNull(message = "Entity id cannot be null")
    private String entityId;

    @NotNull(message = "Domain id for the Entity id cannot be null")
    private String domainId;

    @NotNull(message = "Entity type id cannot be null")
    private String entityTypeId;

    @NotNull(message = "Owner id cannot be null")
    private String ownerId;
    
    private String parentEntityId;
    
    private String name;
    
    private String description;

    private byte[] binaryData;
    private String fullText;
    private Long originalEntityCreationTime;
    private int sharedCount;
    private Long createdTime;
    private Long updatedTime;


    public String getEntityId() {
        return entityId;
    }

    public void setEntityId(String entityId) {
        this.entityId = entityId;
    }

    public String getDomainId() {
        return domainId;
    }

    public void setDomainId(String domainId) {
        this.domainId = domainId;
    }

    public String getEntityTypeId() {
        return entityTypeId;
    }

    public void setEntityTypeId(String entityTypeId) {
        this.entityTypeId = entityTypeId;
    }

    public String getOwnerId() {
        return ownerId;
    }

    public void setOwnerId(String ownerId) {
        this.ownerId = ownerId;
    }

    public String getParentEntityId() {
        return parentEntityId;
    }

    public void setParentEntityId(String parentEntityId) {
        this.parentEntityId = parentEntityId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public byte[] getBinaryData() {
        return binaryData;
    }

    public void setBinaryData(byte[] binaryData) {
        this.binaryData = binaryData;
    }

    public String getFullText() {
        return fullText;
    }

    public void setFullText(String fullText) {
        this.fullText = fullText;
    }

    public int getSharedCount() {
        return sharedCount;
    }

    public void setSharedCount(int sharedCount) {
        this.sharedCount = sharedCount;
    }

    public Long getOriginalEntityCreationTime() {
        return originalEntityCreationTime;
    }

    public void setOriginalEntityCreationTime(Long originalEntityCreationTime) {
        this.originalEntityCreationTime = originalEntityCreationTime;
    }

    public Long getCreatedTime() {
        return createdTime;
    }

    public void setCreatedTime(Long createdTime) {
        this.createdTime = createdTime;
    }

    public Long getUpdatedTime() {
        return updatedTime;
    }

    public void setUpdatedTime(Long updatedTime) {
        this.updatedTime = updatedTime;
    }
}
