package org.apache.custos.sharing.service.core.models;

public class Entity {
    private String entityId;
    private String domainId;
    private String entityTypeId;
    private String ownerId;
    private String parentEntityId;
    private String name;
    private String description;
    private byte[] binaryData;
    private String fullText;
    private double sharedCount = 0;
    private double originalEntityCreationTime;
    private double createdTime;
    private double updatedTime;

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

    public double getSharedCount() {
        return sharedCount;
    }

    public void setSharedCount(double sharedCount) {
        this.sharedCount = sharedCount;
    }

    public double getOriginalEntityCreationTime() {
        return originalEntityCreationTime;
    }

    public void setOriginalEntityCreationTime(double originalEntityCreationTime) {
        this.originalEntityCreationTime = originalEntityCreationTime;
    }

    public double getCreatedTime() {
        return createdTime;
    }

    public void setCreatedTime(double createdTime) {
        this.createdTime = createdTime;
    }

    public double getUpdatedTime() {
        return updatedTime;
    }

    public void setUpdatedTime(double updatedTime) {
        this.updatedTime = updatedTime;
    }
}
