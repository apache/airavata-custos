package org.apache.custos.profile.commons.user.entities;

import javax.persistence.Embeddable;
import java.io.Serializable;
import java.util.Objects;

@Embeddable
public class UserIdentifier implements Serializable {

    private String userId;
    private String gatewayId;

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
    }

    public String getGatewayId() {
        return gatewayId;
    }

    public void setGatewayId(String gatewayId) {
        this.gatewayId = gatewayId;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        UserIdentifier userId1 = (UserIdentifier) o;
        return userId.equals(userId1.userId) &&
                gatewayId.equals(userId1.gatewayId);
    }

    @Override
    public int hashCode() {
        return Objects.hash(userId, gatewayId);
    }
}
