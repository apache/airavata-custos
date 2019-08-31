package org.apache.custos.sharing.service.core.models;

import java.util.List;

public class UserGroup {

    private String groupId;
    private String domainId;
    private String name;
    private String description;
    private String ownerId;
    private GroupType groupType;
    private GroupCardinality groupCardinality;
    private double createdTime;
    private double updatedTime;
    private List<GroupAdmin> groupAdmins;

    public String getGroupId() {
        return groupId;
    }

    public void setGroupId(String groupId) {
        this.groupId = groupId;
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

    public String getOwnerId() {
        return ownerId;
    }

    public void setOwnerId(String ownerId) {
        this.ownerId = ownerId;
    }

    public GroupType getGroupType() {
        return groupType;
    }

    public void setGroupType(GroupType groupType) {
        this.groupType = groupType;
    }

    public GroupCardinality getGroupCardinality() {
        return groupCardinality;
    }

    public void setGroupCardinality(GroupCardinality groupCardinality) {
        this.groupCardinality = groupCardinality;
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

    public List<GroupAdmin> getGroupAdmins() {
        return groupAdmins;
    }

    public void setGroupAdmins(List<GroupAdmin> groupAdmins) {
        this.groupAdmins = groupAdmins;
    }
}
