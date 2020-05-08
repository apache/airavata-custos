package org.apache.custos.integration.tests;

import org.apache.custos.iam.service.AllRoles;
import org.apache.custos.iam.service.RoleRepresentation;
import org.apache.custos.tenant.management.service.CreateTenantResponse;
import org.apache.custos.tenant.management.service.GetTenantResponse;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;
import org.apache.custos.tenant.profile.service.GetAllTenantsResponse;
import org.junit.Assert;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Parameters;
import org.testng.annotations.Test;

import java.io.IOException;

/**
 * This class contains integration tests for Tenant Management
 */
public class TenantManagementTests {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementTests.class);

    private String LOG_SUFFIX = "...........................";


    private TenantManagementClient tenantManagementClient;

    private String serverHost;
    private String serverPort;

    private String clientId;
    private String clientSec;


    @Parameters({"server-host", "server-port", "client-id", "client-sec"})
    @BeforeClass(groups = {"tenant-management"})
    public void setup(String serverHost, String serverPort, String clientId, String clientSec) throws IOException {
        LOGGER.info("Initiating tenant management test cases  " + LOG_SUFFIX);
        this.serverHost = serverHost;
        this.serverPort = serverPort;
        tenantManagementClient = new TenantManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);

    }


    @Test(groups = {"tenant-management"})
    public void createTenant() {
        LOGGER.info("Executing createTenant test case ");


        String[] contants = {
                "custos@airavata.apache.org"
        };

        String[] redirectURI = {"http://localhost:8080/callback"};

        CreateTenantResponse response = tenantManagementClient.registerTenant("Testing tenant",
                "custos@airavata.apache.org",
                "Merry",
                "Jhonson",
                "custos@airavata.apache.org",
                "testuser",
                "12345",
                contants,
                redirectURI,
                "https://test.custos.org",
                "email openid profile org.cilogon.userinfo",
                "test.custos.org",
                "https://test.custos.org",
                "Integration tenant client"
        );

        Assert.assertTrue(response.getIsActivated());
        clientId = response.getClientId();
        clientSec = response.getClientSecret();

    }

    @Test(groups = {"tenant-management"}, dependsOnMethods = {"createTenant"})
    public void getTenant() {
        LOGGER.info("Executing getTenant test case ");
        GetTenantResponse response = tenantManagementClient.getTenant(clientId);
        Assert.assertEquals(response.getClientName(), "Testing tenant");
        Assert.assertEquals(response.getRequesterEmail(), "custos@airavata.apache.org");
        Assert.assertEquals(response.getScope(), "email openid profile org.cilogon.userinfo");
        Assert.assertTrue(response.getRedirectUrisList().contains("http://localhost:8080/callback"));

    }


    @Test(groups = {"tenant-management"}, dependsOnMethods = {"getTenant"})
    public void getChildTenants() {
        LOGGER.info("Executing getChildTenant test case ");
        GetAllTenantsResponse response = tenantManagementClient.getChildTenants(2, 0, "ACTIVE");

        Assert.assertTrue(response.getTenantCount() > 0);


    }


    @Test(groups = {"tenant-management"}, dependsOnMethods = {"getChildTenants"})
    public void updateTenant() {
        LOGGER.info("Executing updateTenant test case");

        String[] contants = {
                "custos@airavata.apache.org"
        };

        String[] redirectURI = {"http://localhost:8080/callback", "http://localhost:8080/callback/updated"};

        GetTenantResponse response = tenantManagementClient.updateTenant(clientId,
                "Testing tenant updated",
                "custos@airavata.apache.org",
                "Merry",
                "Jhonson",
                "custos@airavata.apache.org",
                "testuser",
                "12345",
                contants,
                redirectURI,
                "https://test.custos.org",
                "email openid profile org.cilogon.userinfo",
                "test.custos.org",
                "https://test.custos.org",
                "Integration tenant client"
        );

        Assert.assertEquals(response.getClientName(), "Testing tenant updated");
        Assert.assertTrue(response.getRedirectUrisList().contains("http://localhost:8080/callback/updated"));
    }

    @Test(groups = {"tenant-management"}, dependsOnMethods = {"updateTenant"})
    public void addTenantRoles() throws IOException {
        LOGGER.info("Executing addTenantRoles testcase ");

        TenantManagementClient tenantManagementClient = new TenantManagementClient(this.serverHost,
                Integer.valueOf(this.serverPort), clientId, clientSec);

        RoleRepresentation roleRepresentation = RoleRepresentation.newBuilder()
                .setName("testrole")
                .setDescription("This is testrole").build();
        RoleRepresentation[] roleRepresentations = {roleRepresentation};
        AllRoles allRoles = tenantManagementClient.addTenantRoles(roleRepresentations, false);

        Assert.assertTrue(allRoles.getRolesCount() > 0);

        boolean isContain = false;

        for (RoleRepresentation representation : allRoles.getRolesList()) {

            if (representation.getName().equals("testrole")) {
                isContain = true;
            }
        }

        Assert.assertTrue(isContain);

        AllRoles clientRoles = tenantManagementClient.addTenantRoles(roleRepresentations, true);
        boolean isClientRoleContain = false;

        for (RoleRepresentation representation : clientRoles.getRolesList()) {

            if (representation.getName().equals("testrole")) {
                isClientRoleContain = true;
            }
        }
        Assert.assertTrue(isClientRoleContain);

    }


    @Test(groups = {"tenant-management"}, dependsOnMethods = {"addTenantRoles"})
    public void deleteTenant() {
        LOGGER.info("Executing delete tenant test");
        tenantManagementClient.deleteTenant(clientId);

    }


    @AfterClass(groups = {"tenant-management"})
    void cleanup() {
        LOGGER.info("Completing tenant management tests " + LOG_SUFFIX);
        clientId = null;
        clientSec = null;
        serverPort = null;
        serverHost = null;
    }

}
