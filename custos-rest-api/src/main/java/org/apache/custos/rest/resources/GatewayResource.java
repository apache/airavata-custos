package org.apache.custos.rest.resources;

public class GatewayResource {
    private String custosInternalGatewayId; // optional
    private String gatewayId; // required
    private String gatewayApprovalStatus; // required
    private String gatewayName; // optional
    private String domain; // optional
    private String emailAddress; // optional
    private String gatewayAcronym; // optional
    private String gatewayURL; // optional
    private String gatewayPublicAbstract; // optional
    private String reviewProposalDescription; // optional
    private String gatewayAdminFirstName; // optional
    private String gatewayAdminLastName; // optional
    private String gatewayAdminEmail; // optional
    private String identityServerUserName; // optional
    private String identityServerPasswordToken; // optional
    private String declinedReason; // optional
    private String oauthClientId; // optional
    private String oauthClientSecret; // optional
    private long requestCreationTime; // optional
    private String requesterUsername; // optional

    public String getCustosInternalGatewayId() {
        return custosInternalGatewayId;
    }

    public void setCustosInternalGatewayId(String custosInternalGatewayId) {
        this.custosInternalGatewayId = custosInternalGatewayId;
    }

    public String getGatewayId() {
        return gatewayId;
    }

    public void setGatewayId(String gatewayId) {
        this.gatewayId = gatewayId;
    }

    public String getGatewayApprovalStatus() {
        return gatewayApprovalStatus;
    }

    public void setGatewayApprovalStatus(String gatewayApprovalStatus) {
        this.gatewayApprovalStatus = gatewayApprovalStatus;
    }

    public String getGatewayName() {
        return gatewayName;
    }

    public void setGatewayName(String gatewayName) {
        this.gatewayName = gatewayName;
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public String getEmailAddress() {
        return emailAddress;
    }

    public void setEmailAddress(String emailAddress) {
        this.emailAddress = emailAddress;
    }

    public String getGatewayAcronym() {
        return gatewayAcronym;
    }

    public void setGatewayAcronym(String gatewayAcronym) {
        this.gatewayAcronym = gatewayAcronym;
    }

    public String getGatewayURL() {
        return gatewayURL;
    }

    public void setGatewayURL(String gatewayURL) {
        this.gatewayURL = gatewayURL;
    }

    public String getGatewayPublicAbstract() {
        return gatewayPublicAbstract;
    }

    public void setGatewayPublicAbstract(String gatewayPublicAbstract) {
        this.gatewayPublicAbstract = gatewayPublicAbstract;
    }

    public String getReviewProposalDescription() {
        return reviewProposalDescription;
    }

    public void setReviewProposalDescription(String reviewProposalDescription) {
        this.reviewProposalDescription = reviewProposalDescription;
    }

    public String getGatewayAdminFirstName() {
        return gatewayAdminFirstName;
    }

    public void setGatewayAdminFirstName(String gatewayAdminFirstName) {
        this.gatewayAdminFirstName = gatewayAdminFirstName;
    }

    public String getGatewayAdminLastName() {
        return gatewayAdminLastName;
    }

    public void setGatewayAdminLastName(String gatewayAdminLastName) {
        this.gatewayAdminLastName = gatewayAdminLastName;
    }

    public String getGatewayAdminEmail() {
        return gatewayAdminEmail;
    }

    public void setGatewayAdminEmail(String gatewayAdminEmail) {
        this.gatewayAdminEmail = gatewayAdminEmail;
    }

    public String getIdentityServerUserName() {
        return identityServerUserName;
    }

    public void setIdentityServerUserName(String identityServerUserName) {
        this.identityServerUserName = identityServerUserName;
    }

    public String getIdentityServerPasswordToken() {
        return identityServerPasswordToken;
    }

    public void setIdentityServerPasswordToken(String identityServerPasswordToken) {
        this.identityServerPasswordToken = identityServerPasswordToken;
    }

    public String getDeclinedReason() {
        return declinedReason;
    }

    public void setDeclinedReason(String declinedReason) {
        this.declinedReason = declinedReason;
    }

    public String getOauthClientId() {
        return oauthClientId;
    }

    public void setOauthClientId(String oauthClientId) {
        this.oauthClientId = oauthClientId;
    }

    public String getOauthClientSecret() {
        return oauthClientSecret;
    }

    public void setOauthClientSecret(String oauthClientSecret) {
        this.oauthClientSecret = oauthClientSecret;
    }

    public long getRequestCreationTime() {
        return requestCreationTime;
    }

    public void setRequestCreationTime(long requestCreationTime) {
        this.requestCreationTime = requestCreationTime;
    }

    public String getRequesterUsername() {
        return requesterUsername;
    }

    public void setRequesterUsername(String requesterUsername) {
        this.requesterUsername = requesterUsername;
    }
}
