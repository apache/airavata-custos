package org.apache.custos.sharing.service.core.models;

import javax.validation.constraints.NotNull;
import java.util.ArrayList;
import java.util.List;

public class UserGroup {

    @NotNull(message = "Group id cannot be null")
    private String groupId;

    @NotNull(message = "Domain id from the group cannot be null")
    private String domainId;

    @NotNull(message = "Name from the group cannot be null")
    private String name;
    
    private String description;

    @NotNull(message = "OwnerId of the group cannot be null")
    private String ownerId;

    @NotNull(message = "Group type cannot be null")
    private GroupType groupType;
    
    private GroupCardinality groupCardinality;
    
    private Long createdTime;
    
    private Long updatedTime;

    private List<GroupAdmin> groupAdmins = new ArrayList<>();

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

    public List<GroupAdmin> getGroupAdmins() {
        return groupAdmins;
    }

    public void setGroupAdmins(List<GroupAdmin> groupAdmins) {
        this.groupAdmins = groupAdmins;
    }
}
