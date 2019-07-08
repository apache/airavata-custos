package org.apache.custos.authentication.handler;

import org.apache.custos.authentication.cpi.CustosAuthenticationService;
import org.apache.custos.authentication.cpi.exception.CustosAuthenticationServiceException;
import org.apache.custos.commons.model.security.UserInfo;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.error.AuthenticationException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.security.manager.CustosSecurityManager;
import org.apache.custos.security.manager.SecurityManagerFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CustosAuthenticationHandler implements CustosAuthenticationService.Iface {
    private final static Logger logger = LoggerFactory.getLogger(CustosAuthenticationHandler.class);

    @Override
    public boolean isUserAuthenticated(AuthzToken authzToken) throws AuthenticationException {
        try {
            CustosSecurityManager securityManager = SecurityManagerFactory.getSecurityManager();
            boolean isAuth = securityManager.isUserAuthenticated(authzToken);
                if (isAuth) {
                    logger.info("User" + authzToken.getClaimsMap().get(Constants.USER_NAME) + "in gateway" + authzToken.getClaimsMap().get(Constants.GATEWAY_ID) + "is authenticated");
                    return isAuth;
                }
                else{
                    throw new AuthenticationException("User is not authenticated.");
                }
        } catch (CustosSecurityException e) {
            logger.error(e.getMessage(), e);
            throw new AuthenticationException("Error in authenticating.");
        }
    }
    @Override
    public UserInfo getUserInfoFromAuthzToken(AuthzToken authzToken) throws CustosAuthenticationServiceException {
        try{
            CustosSecurityManager securityManager = SecurityManagerFactory.getSecurityManager();
            UserInfo userInfo = securityManager.getUserInfoFromAuthzToken(authzToken);
            return userInfo;
        }catch (CustosSecurityException e){
            logger.error(e.getMessage(), e);
            throw new CustosAuthenticationServiceException("Could not retrieve user info");
        }
    }
    @Override
    public AuthzToken getUserManagementServiceAccountAuthzToken(AuthzToken authzToken, String gatewayId) throws CustosAuthenticationServiceException {
        try{
            CustosSecurityManager securityManager = SecurityManagerFactory.getSecurityManager();
            AuthzToken managementServiceAccountAuthzToken = securityManager.getUserManagementServiceAccountAuthzToken(authzToken, gatewayId);
            return managementServiceAccountAuthzToken;
        }catch (CustosSecurityException e){
            logger.error(e.getMessage(), e);
            throw new CustosAuthenticationServiceException("Could get user management account authz token");
        }
    }
}
