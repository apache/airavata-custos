package org.apache.custos.sharing.service.core.models;

public class PermissionType {

    private String permissionTypeId;
    private String domainId;
    private String name;
    private String description;
    private double createdTime;
    private double updatedTime;

    public String getPermissionTypeId() {
        return permissionTypeId;
    }

    public void setPermissionTypeId(String permissionTypeId) {
        this.permissionTypeId = permissionTypeId;
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
