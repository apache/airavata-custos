package org.apache.custos.security.service;

import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.error.AuthorizationException;
import org.apache.custos.security.manager.CustosSecurityManager;
import org.apache.custos.security.manager.SecurityManagerFactory;
import org.apache.custos.security.model.AuthzToken;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

public class SecurityServiceHandler {

    private final static Logger logger = LoggerFactory.getLogger(SecurityServiceHandler.class);

    public boolean authorize(AuthzToken authzToken, Map<String, String> metaData) throws AuthorizationException {
        try {
            CustosSecurityManager securityManager = SecurityManagerFactory.getSecurityManager();
            boolean isAuthz = securityManager.isUserAuthorized(authzToken, metaData);
            if (!isAuthz) {
                throw new AuthorizationException("User is not authenticated or authorized.");
            }
        } catch (CustosSecurityException e) {
            logger.error(e.getMessage(), e);
            throw new AuthorizationException("Error in authenticating or authorizing user.");
        }
        return true;
    }
}
