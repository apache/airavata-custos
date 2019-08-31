package org.apache.custos.sharing.service.core.models;

public class GroupMembership {

    private String parentId;
    private String childId;
    private String domainId;
    private GroupChildType childType;
    private double createdTime;
    private double updatedTime;

    public String getParentId() {
        return parentId;
    }

    public void setParentId(String parentId) {
        this.parentId = parentId;
    }

    public String getChildId() {
        return childId;
    }

    public void setChildId(String childId) {
        this.childId = childId;
    }

    public String getDomainId() {
        return domainId;
    }

    public void setDomainId(String domainId) {
        this.domainId = domainId;
    }

    public GroupChildType getChildType() {
        return childType;
    }

    public void setChildType(GroupChildType childType) {
        this.childType = childType;
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
