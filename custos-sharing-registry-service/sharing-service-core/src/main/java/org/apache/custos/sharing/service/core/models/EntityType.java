package org.apache.custos.sharing.service.core.models;

import javax.validation.constraints.NotNull;

public class EntityType {

    @NotNull(message = "Entity type id cannot be null")
    private String entityTypeId;

    @NotNull(message = "Domain id for entity type cannot be null")
    private String domainId;

    @NotNull(message = "Name for entity type cannot be null")
    private String name;

    private String description;

    private Long createdTime;

    private Long updatedTime;

    public String getEntityTypeId() {
        return entityTypeId;
    }

    public void setEntityTypeId(String entityTypeId) {
        this.entityTypeId = entityTypeId;
    }

    public String getDomainId() {
        return domainId;
    }

    public void setDomainId(String domainId) {
        this.domainId = domainId;
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
