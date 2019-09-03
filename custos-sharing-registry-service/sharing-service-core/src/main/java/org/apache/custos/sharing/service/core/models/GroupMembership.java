package org.apache.custos.sharing.service.core.models;

public class GroupMembership {

    private String parentId;
    private String childId;
    private String domainId;
    private GroupChildType childType;
    private Long createdTime;
    private Long updatedTime;

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
