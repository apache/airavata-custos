package org.apache.custos.integration.tests;

import org.apache.custos.identity.management.client.IdentityManagementClient;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;
import org.apache.custos.user.management.client.UserManagementClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * This class contains AgentManagement tests
 */
public class AgentManagementTests {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentManagementTests.class);

    private String LOG_SUFFIX = "...........................";

    private UserManagementClient userManagementClient;
    private IdentityManagementClient identityManagementClient;
    private TenantManagementClient tenantManagementClient;

    private String username;


}
