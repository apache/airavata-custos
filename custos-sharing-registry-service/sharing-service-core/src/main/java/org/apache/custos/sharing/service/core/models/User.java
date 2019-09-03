package org.apache.custos.sharing.service.core.models;

import javax.validation.constraints.NotNull;
import java.nio.ByteBuffer;

public class User {

    @NotNull(message = "User id cannot be null")
    private String userId;

    @NotNull(message = "Domain to which user belongs cannot be null")
    private String domainId;

    @NotNull(message = "username cannot be null")
    private String userName;

    private String firstName;

    private String lastName;

    private String email;

    private ByteBuffer icon;

    private Long createdTime;
    
    private Long updatedTime;

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
    }

    public String getDomainId() {
        return domainId;
    }

    public void setDomainId(String domainId) {
        this.domainId = domainId;
    }

    public String getUserName() {
        return userName;
    }

    public void setUserName(String userName) {
        this.userName = userName;
    }

    public String getFirstName() {
        return firstName;
    }

    public void setFirstName(String firstName) {
        this.firstName = firstName;
    }

    public String getLastName() {
        return lastName;
    }

    public void setLastName(String lastName) {
        this.lastName = lastName;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public ByteBuffer getIcon() {
        return icon;
    }

    public void setIcon(ByteBuffer icon) {
        this.icon = icon;
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
