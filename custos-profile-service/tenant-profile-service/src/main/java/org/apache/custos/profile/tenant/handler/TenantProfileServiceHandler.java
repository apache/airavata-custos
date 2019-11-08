/**
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
 */
package org.apache.custos.profile.tenant.handler;

import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.profile.commons.tenant.entities.GatewayEntity;
import org.apache.custos.profile.model.workspace.Gateway;
import org.apache.custos.profile.model.workspace.GatewayApprovalStatus;
import org.apache.custos.profile.commons.repositories.TenantProfileRepository;
import org.apache.custos.profile.tenant.cpi.TenantProfileService;
import org.apache.custos.profile.tenant.cpi.exception.TenantProfileServiceException;
import org.apache.custos.profile.tenant.cpi.profile_tenant_cpiConstants;
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

    public TenantProfileServiceHandler() {
        logger.debug("Initializing TenantProfileServiceHandler");
        this.tenantProfileRepository = new TenantProfileRepository(Gateway.class, GatewayEntity.class);
    }

    @Override
    public String getAPIVersion() throws TException {
        return profile_tenant_cpiConstants.TENANT_PROFILE_CPI_VERSION;
    }

    @Override
    public Gateway addGateway(AuthzToken authzToken, Gateway gateway) throws TenantProfileServiceException, TException {
        try {
            // Assign UUID to gateway
            gateway.setCustosInternalGatewayId(UUID.randomUUID().toString());
            if (!checkDuplicateGateway(authzToken, gateway)) {
                gateway = tenantProfileRepository.create(gateway);
                if (gateway != null) {
                    logger.info("Added Custos Gateway with Id: " + gateway.getGatewayId());
                    // TODO: approval status will be set to approved only if gateway object has approved status
                    if (gateway.getGatewayApprovalStatus().equals(GatewayApprovalStatus.APPROVED)) {
                        logger.info("Gateway with ID: {}, is now APPROVED, returning", gateway.getGatewayId());
                        return gateway;
                    }
                    else {
                        throw new Exception("Gateway not approved");
                    }
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
    public Gateway updateGateway(AuthzToken authzToken, Gateway updatedGateway) throws TenantProfileServiceException, TException {
        try {
            // if admin password token changes then copy the admin password and store under this gateway id and then update the admin password token
            Gateway existingGateway = tenantProfileRepository.getGateway(updatedGateway.getCustosInternalGatewayId());
            if (tenantProfileRepository.update(updatedGateway) != null) {
                logger.debug("Updated gateway-profile with ID: " + updatedGateway.getGatewayId());
                return existingGateway;
            } else {
                throw new TenantProfileServiceException("Could not update the gateway");
            }
        } catch (Exception ex) {
            logger.error("Error updating gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error updating gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    public Gateway getGateway(AuthzToken authzToken, String airavataInternalGatewayId) throws TenantProfileServiceException, TException {
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

    @Override
    public Gateway getGatewayUsingGatewayId(AuthzToken authzToken, String gatewayId) throws TenantProfileServiceException, TException {
        try {
            Gateway gateway = tenantProfileRepository.getGatewayUsingGatewayId(gatewayId);
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
    public Gateway deleteGateway(AuthzToken authzToken, String custosInternalGatewayId, String gatewayId) throws TenantProfileServiceException, TException {
        try {
            logger.debug("Deleting Custos gateway-profile with ID: " + gatewayId + "Internal ID: " + custosInternalGatewayId);
            boolean deleteSuccess = tenantProfileRepository.delete(custosInternalGatewayId);
            if (deleteSuccess) {
                return new Gateway(gatewayId, GatewayApprovalStatus.DEACTIVATED);
            }
            else {
                throw new Exception("Error in deleting");
            }
        } catch (Exception ex) {
            logger.error("Error deleting gateway-profile, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error deleting gateway-profile, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    public List<Gateway> getAllGateways(AuthzToken authzToken) throws TenantProfileServiceException, TException {
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
    public boolean isGatewayExist(AuthzToken authzToken, String gatewayId) throws TenantProfileServiceException, TException {
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
    public List<Gateway> getAllGatewaysForUser(AuthzToken authzToken, String requesterUsername) throws TenantProfileServiceException, TException {
        try {
            return tenantProfileRepository.getAllGatewaysForUser(requesterUsername);
        } catch (Exception ex) {
            logger.error("Error getting user's gateway-profiles, reason: " + ex.getMessage(), ex);
            TenantProfileServiceException exception = new TenantProfileServiceException();
            exception.setMessage("Error getting user's gateway-profiles, reason: " + ex.getMessage());
            throw exception;
        }
    }

    @Override
    public boolean checkDuplicateGateway(AuthzToken authzToken, Gateway gateway) throws TenantProfileServiceException {
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
}
