/*
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
*/
package org.apache.custos.profile.handlers;

import org.apache.custos.commons.model.error.AuthorizationException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.profile.model.workspace.Gateway;
import org.apache.custos.profile.model.workspace.GatewayApprovalStatus;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.profile.tenant.cpi.exception.TenantProfileServiceException;
import org.apache.custos.profile.tenant.cpi.profile_tenant_cpiConstants;
import org.apache.custos.security.interceptor.SecurityCheck;
import org.apache.custos.profile.commons.tenant.entities.GatewayEntity;
import org.apache.custos.profile.tenant.core.repositories.TenantProfileRepository;
import org.apache.thrift.TException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.UUID;

/**
 * Created by goshenoy on 3/6/17.
 */
public class TenantProfileServiceHandler implements TenantProfileService.Iface {

    private final static Logger logger = LoggerFactory.getLogger(TenantProfileServiceHandler.class);

    private TenantProfileRepository tenantProfileRepository;
    //private DBEventPublisherUtils dbEventPublisherUtils = new DBEventPublisherUtils(DBEventService.TENANT);

    public TenantProfileServiceHandler() {
        logger.debug("Initializing TenantProfileServiceHandler");
        this.tenantProfileRepository = new TenantProfileRepository(Gateway.class, GatewayEntity.class);
    }

    @Override
    public String getAPIVersion() throws TenantProfileServiceException, TException {
        return profile_tenant_cpiConstants.TENANT_PROFILE_CPI_VERSION;
    }

    @Override
    @SecurityCheck
    public Gateway addGateway(AuthzToken authzToken, Gateway gateway) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            // Assign UUID to gateway
            gateway.setCustosInternalGatewayId(UUID.randomUUID().toString());
            if (!checkDuplicateGateway(gateway)) {
                // If admin password, copy it in the credential store under the requested gateway's gatewayId
//                if (gateway.getIdentityServerPasswordToken() != null) {
//                    copyAdminPasswordToGateway(authzToken, gateway);
//                }
                gateway = tenantProfileRepository.create(gateway);
                if (gateway != null) {
                    logger.info("Added Custos Gateway with Id: " + gateway.getGatewayId());
                    // replicate tenant at end-places only if status is APPROVED
                    if (gateway.getGatewayApprovalStatus().equals(GatewayApprovalStatus.APPROVED)) {
                        logger.info("Gateway with ID: {}, is now APPROVED, replicating to subscribers.", gateway.getGatewayId());
                        return gateway;
                        //dbEventPublisherUtils.publish(EntityType.TENANT, CrudType.CREATE, gateway);
                    }
                    else {
                        throw new Exception("Gateway status is not approved");
                    }
//                    // return internal id
//                    return gateway.getAiravataInternalGatewayId();
                } else {
                    throw new Exception("Gateway object is null.");
                }
            }
            else {
                throw new TenantProfileServiceException("An approved Gateway already exists with the same GatewayId, Name or URL");
            }
        } catch (Exception ex) {
            logger.error("Error adding gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error adding gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public Gateway updateGateway(AuthzToken authzToken, Gateway updatedGateway) throws TenantProfileServiceException, AuthorizationException, TException {
        try {

//            // if admin password token changes then copy the admin password and store under this gateway id and then update the admin password token
//            Gateway existingGateway = tenantProfileRepository.getGateway(updatedGateway.getAiravataInternalGatewayId());
//            if (updatedGateway.getIdentityServerPasswordToken() != null
//                    && (existingGateway.getIdentityServerPasswordToken() == null
//                        || !existingGateway.getIdentityServerPasswordToken().equals(updatedGateway.getIdentityServerPasswordToken()))) {
//                copyAdminPasswordToGateway(authzToken, updatedGateway);
//            }

            if (tenantProfileRepository.update(updatedGateway) != null) {
                logger.debug("Updated gateway-profile with ID: " + updatedGateway.getGatewayId());
                // replicate tenant at end-places
//                dbEventPublisherUtils.publish(EntityType.TENANT, CrudType.UPDATE, updatedGateway);
                return updatedGateway;
            } else {
                throw new TException("Could not update the gateway");
            }
        } catch (Exception ex) {
            logger.error("Error updating gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error updating gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public Gateway getGateway(AuthzToken authzToken, String airavataInternalGatewayId) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            Gateway gateway = tenantProfileRepository.getGateway(airavataInternalGatewayId);
            if (gateway == null) {
                throw new Exception("Could not find Gateway with internal ID: " + airavataInternalGatewayId);
            }
            return gateway;
        } catch (Exception ex) {
            logger.error("Error getting gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error getting gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    //TODO: check why this was not included, add authToken
    @Override
    @SecurityCheck
    public Gateway getGatewayUsingGatewayId(String gatewayId) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            Gateway gateway = tenantProfileRepository.getGatewayUsingId(gatewayId);
            if (gateway == null) {
                throw new Exception("Could not find Gateway with ID: " + gatewayId);
            }
            return gateway;
        } catch (Exception ex) {
            logger.error("Error getting gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error getting gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public Gateway deleteGateway(AuthzToken authzToken, String custosInternalGatewayId, String gatewayId) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            logger.debug("Deleting Custos gateway-profile with ID: " + gatewayId + "Internal ID: " + custosInternalGatewayId);
            boolean deleteSuccess = tenantProfileRepository.delete(custosInternalGatewayId);
            if (deleteSuccess) {
                // delete tenant at end-places
//                dbEventPublisherUtils.publish(EntityType.TENANT, CrudType.DELETE,
//                        // pass along gateway datamodel, with correct gatewayId;
//                        // approvalstatus is not used for delete, hence set dummy value
//                        new Gateway(gatewayId, GatewayApprovalStatus.DEACTIVATED));
                return new Gateway(gatewayId, GatewayApprovalStatus.DEACTIVATED);
            }
            else {
                throw new Exception("Deleting gateway failed");
            }
        } catch (Exception ex) {
            logger.error("Error deleting gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error deleting gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public List<Gateway> getAllGateways(AuthzToken authzToken) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            return tenantProfileRepository.getAllGateways();
        } catch (Exception ex) {
            logger.error("Error getting all gateway-profiles, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error getting all gateway-profiles, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public boolean isGatewayExist(AuthzToken authzToken, String gatewayId) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            Gateway gateway = tenantProfileRepository.getGateway(gatewayId);
            return (gateway != null);
        } catch (Exception ex) {
            logger.error("Error checking if gateway-profile exists, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error checking if gateway-profile exists, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    @SecurityCheck
    public List<Gateway> getAllGatewaysForUser(AuthzToken authzToken, String requesterUsername) throws TenantProfileServiceException, AuthorizationException, TException {
        try {
            return tenantProfileRepository.getAllGatewaysForUser(requesterUsername);
        } catch (Exception ex) {
            logger.error("Error getting user's gateway-profiles, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error getting user's gateway-profiles, reason: " + ex.getMessage());
            throw exception;
        }
    }

    private boolean checkDuplicateGateway(Gateway gateway) throws TenantProfileServiceException {
        try {
            Gateway duplicateGateway = tenantProfileRepository.getDuplicateGateway(gateway.getGatewayId(), gateway.getGatewayName(), gateway.getGatewayURL());
            return duplicateGateway != null;
        } catch (Exception ex) {
            logger.error("Error checking if duplicate gateway-profile exists, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error checking if duplicate gateway-profiles exists, reason: " + ex.getMessage());
            throw exception;
        }
    }

//    // admin passwords are stored in credential store in the super portal gateway and need to be
//    // copied to a credential that is stored in the requested/newly created gateway
//    private void copyAdminPasswordToGateway(AuthzToken authzToken, Gateway gateway) throws TException, ApplicationSettingsException {
//        CredentialStoreService.Client csClient = getCredentialStoreServiceClient();
//        try {
//            String requestGatewayId = authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
//            PasswordCredential adminPasswordCredential = csClient.getPasswordCredential(gateway.getIdentityServerPasswordToken(), requestGatewayId);
//            adminPasswordCredential.setGatewayId(gateway.getGatewayId());
//            String newAdminPasswordCredentialToken = csClient.addPasswordCredential(adminPasswordCredential);
//            gateway.setIdentityServerPasswordToken(newAdminPasswordCredentialToken);
//        } finally {
//            if (csClient.getInputProtocol().getTransport().isOpen()) {
//                csClient.getInputProtocol().getTransport().close();
//            }
//            if (csClient.getOutputProtocol().getTransport().isOpen()) {
//                csClient.getOutputProtocol().getTransport().close();
//            }
//        }
//    }

//    private CredentialStoreService.Client getCredentialStoreServiceClient() throws TException, ApplicationSettingsException {
//        final int serverPort = Integer.parseInt(ServerSettings.getCredentialStoreServerPort());
//        final String serverHost = ServerSettings.getCredentialStoreServerHost();
//        try {
//            return CredentialStoreClientFactory.createAiravataCSClient(serverHost, serverPort);
//        } catch (CredentialStoreException e) {
//            throw new TException("Unable to create credential store client...", e);
//        }
//    }
}
