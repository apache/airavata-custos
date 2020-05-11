package org.apache.custos.integration.tests;

import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import org.apache.custos.agent.management.client.AgentManagementClient;
import org.apache.custos.agent.management.service.AgentRegistrationResponse;
import org.apache.custos.iam.service.Agent;
import org.apache.custos.iam.service.OperationStatus;
import org.apache.custos.iam.service.UserAttribute;
import org.apache.custos.identity.management.client.IdentityManagementClient;
import org.junit.Assert;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Parameters;
import org.testng.annotations.Test;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * This class contains AgentManagement tests
 */
public class AgentManagementTests {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentManagementTests.class);

    private String LOG_SUFFIX = "...........................";

    private AgentManagementClient agentManagementClient;
    private IdentityManagementClient identityManagementClient;


    private String clientId;

    private String adminToken;

    private String registeredAgentId;

    private String agentId;

    private String agentSec;

    @Parameters({"server-host", "server-port", "client-id", "client-sec", "admin-username", "admin-password"})
    @BeforeClass(groups = {"agent-management"})
    public void setup(String serverHost, String serverPort, String clientId, String clientSec,
                      String adminUsername, String adminPassword) throws IOException {
        LOGGER.info("Initiating agent management test cases  " + LOG_SUFFIX);
        agentManagementClient = new AgentManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);
        identityManagementClient = new IdentityManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);
        Struct struct = identityManagementClient.getToken(null,
                null, adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");
        this.adminToken = value.getStringValue();
        this.clientId = clientId;
    }


    @Test(groups = {"agent-management"})
    public void enableAgents() {
        LOGGER.info("Executing enable agents test case ");
        OperationStatus status = agentManagementClient.enableAgents(adminToken);
        Assert.assertTrue(status.getStatus());
    }


    @Test(groups = {"agent-management"}, dependsOnMethods = {"enableAgents"})
    public void registerAgent() {
        LOGGER.info("Executing register agent test case ");
        registeredAgentId = getAlphaNumericString(7);
        List<String> arrayList = new ArrayList<>();
        arrayList.add("true");
        UserAttribute attribute = UserAttribute.newBuilder()
                .setKey("server_access")
                .addAllValues(arrayList)
                .build();
        UserAttribute[] attributes = {attribute};
        String[] realmRoles = {};
        AgentRegistrationResponse registrationResponse = agentManagementClient.
                registerAndEnableAgent(adminToken, registeredAgentId, realmRoles, attributes);
        agentId = registrationResponse.getId();
        Assert.assertEquals(agentId, registeredAgentId);
        agentSec = registrationResponse.getSecret();
    }

    @Test(groups = {"agent-management"}, dependsOnMethods = {"registerAgent"})
    public void getAgent() {
        LOGGER.info("Executing get agent test case ");
        Agent agent = agentManagementClient.getAgent(adminToken, agentId);
        Assert.assertTrue(agent.getId().equalsIgnoreCase(agentId));
        Assert.assertTrue(!agent.getAttributesList().isEmpty());
        List<UserAttribute> attribute = agent.getAttributesList();
        boolean attrMatched = false;

        for (UserAttribute userAttribute : attribute) {
            if (userAttribute.getKey().equals("server_access") && userAttribute.getValuesList().contains("true")) {
                attrMatched = true;
            }
        }
        Assert.assertTrue(attrMatched);

    }

    @Test(groups = {"agent-management"}, dependsOnMethods = {"getAgent"})
    public void disableAgent() {
        LOGGER.info("Executing get agent test case ");
        OperationStatus status = agentManagementClient.disableAgent(adminToken, agentId);
        Assert.assertTrue(status.getStatus());
    }

    @Test(groups = {"agent-management"}, dependsOnMethods = {"disableAgent"})
    public void enableAgent() {
        LOGGER.info("Executing enable agent test case ");
        OperationStatus status = agentManagementClient.enableAgent(adminToken, agentId);
        Assert.assertTrue(status.getStatus());
    }


    @Test(groups = {"agent-management"}, dependsOnMethods = {"enableAgent"})
    public void addAgentAttribute() {
        LOGGER.info("Executing add agent attribute test case ");


        List<String> arrayList = new ArrayList<>();
        arrayList.add("testing");
        UserAttribute attribute = UserAttribute.newBuilder()
                .setKey("test-atr")
                .addAllValues(arrayList)
                .build();
        UserAttribute[] attributes = {attribute};

        String[] agents = {agentId};

        OperationStatus status = agentManagementClient.addAgentAttributes(adminToken, agents, attributes);
        Assert.assertTrue(status.getStatus());

        Agent agent = agentManagementClient.getAgent(adminToken, agentId);

        List<UserAttribute> atrs = agent.getAttributesList();
        boolean attrMatched = false;

        for (UserAttribute userAttribute : atrs) {
            if (userAttribute.getKey().equals("test-atr") && userAttribute.getValuesList().contains("testing")) {
                attrMatched = true;
            }
        }

        Assert.assertTrue(attrMatched);

    }


    @Test(groups = {"agent-management"}, dependsOnMethods = {"addAgentAttribute"})
    public void deleteAgentAttribute() {
        LOGGER.info("Executing delete agent attribute test case ");

        List<String> arrayList = new ArrayList<>();
        arrayList.add("testing");
        UserAttribute attribute = UserAttribute.newBuilder()
                .setKey("test-atr")
                .addAllValues(arrayList)
                .build();
        UserAttribute[] attributes = {attribute};

        String[] agents = {agentId};

        OperationStatus status = agentManagementClient.deleteAgentAttributes(adminToken, agents, attributes);
        Assert.assertTrue(status.getStatus());

        Agent agent = agentManagementClient.getAgent(adminToken, agentId);

        List<UserAttribute> atrs = agent.getAttributesList();
        boolean attrMatched = true;

        for (UserAttribute userAttribute : atrs) {
            if (userAttribute.getKey().equals("test-atr") && userAttribute.getValuesList().contains("testing")) {
                attrMatched = false;
            }
        }

        Assert.assertTrue(attrMatched);

    }

    @Test(groups = {"agent-management"}, dependsOnMethods = {"deleteAgentAttribute"})
    public void getAgentToken() {
        LOGGER.info("Executing get  agent token test case ");
        Struct struct = identityManagementClient.getAgentToken(clientId, agentId,
                agentSec, "client_credentials", null);
        Value value = struct.getFieldsMap().get("access_token");
        Assert.assertTrue(!value.getStringValue().isEmpty());
    }

    @Test(groups = {"agent-management"}, dependsOnMethods = {"getAgentToken"})
    public void deleteAgent() {
        LOGGER.info("Executing delete  agent token test case ");
        OperationStatus status = agentManagementClient.deleteAgent(adminToken, agentId);
        Assert.assertTrue(status.getStatus());
    }

    @AfterClass
    public void cleanup() {
        LOGGER.info("Completing agent management tests " + LOG_SUFFIX);
        agentManagementClient = null;
        identityManagementClient = null;
        clientId = null;
        adminToken = null;
        registeredAgentId = null;
        agentId = null;
        agentSec = null;

    }

    static String getAlphaNumericString(int n) {

        // chose a Character random from this String
        String AlphaNumericString = "abcdefghijklmnopqrstuvxyz"
                + "0123456789";

        // create StringBuffer size of AlphaNumericString
        StringBuilder sb = new StringBuilder(n);

        for (int i = 0; i < n; i++) {

            // generate a random number between
            // 0 to AlphaNumericString variable length
            int index
                    = (int) (AlphaNumericString.length()
                    * Math.random());

            // add Character one by one in end of sb
            sb.append(AlphaNumericString
                    .charAt(index));
        }
        return sb.toString();
    }

}
