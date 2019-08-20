import mockit.*;
import mockit.integration.junit4.JMockit;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.model.security.UserInfo;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.profile.model.user.UserProfile;
import org.apache.custos.profile.user.cpi.exception.UserProfileServiceException;
import org.apache.custos.profile.user.handler.UserProfileServiceHandler;
import org.apache.custos.security.manager.CustosSecurityManager;
import org.apache.custos.security.manager.SecurityManagerFactory;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.util.HashMap;
import java.util.List;

import static org.junit.Assert.*;

@RunWith(JMockit.class)
public class TestUserProfileServiceHandler {

    private static String GATEWAY_ID = "test-gateway";
    private static String USER_NAME = "default-user";
    private static AuthzToken authzToken = new AuthzToken();
    private static String NEW_USER_ID = "test-username";
    private int NUMBER_OF_USERS = 1;
    @Tested
    private static UserProfileServiceHandler userProfileServiceHandler;
    @Mocked
    private static CustosSecurityManager mockedSecurityManager;

    @Before
    public void setUp() throws CustosSecurityException {
        HashMap<String,String> map = new HashMap<>();
        map.put(Constants.GATEWAY_ID, GATEWAY_ID);
        map.put(Constants.USER_NAME, USER_NAME);
        authzToken.setClaimsMap(map);
        authzToken.setAccessToken("access token");
    }
    @After
    public void cleanUp() {

    }
    @Test
    public void testInitializeUserProfileWhenUserProfileDoesNotExist(@Mocked CustosSecurityManager mockedSecurityManager) throws UserProfileServiceException, CustosSecurityException {
        UserInfo userInfo = new UserInfo();
        userInfo.setUsername("test-username");
        userInfo.setEmailAddress("test@test.com");
        userInfo.setFirstName("test-first-name");
        userInfo.setLastName("test-last-name");
        new MockUp<SecurityManagerFactory>() {
            @Mock
            public CustosSecurityManager getSecurityManager(){return mockedSecurityManager;};
        };
        new Expectations() {{ mockedSecurityManager.getUserInfoFromAuthzToken(authzToken); result = userInfo;}};

        UserProfile createdUserProfile = userProfileServiceHandler.initializeUserProfile(authzToken);
        assertNotNull(createdUserProfile);
        assertEquals(userInfo.getUsername().toLowerCase(), createdUserProfile.getUserId());
        assertEquals(GATEWAY_ID,createdUserProfile.getGatewayId());
    }
    @Test
    public void testInitializeUserProfileWhenUserProfileExists(@Mocked CustosSecurityManager mockedSecurityManager) throws UserProfileServiceException, CustosSecurityException{
        UserInfo userInfo = new UserInfo();
        userInfo.setUsername("test-username");
        userInfo.setEmailAddress("some-other-test@test.com");
        userInfo.setFirstName("someother-test-first-name");
        userInfo.setLastName("someother-test-last-name");
        String old_email_address = "test@test.com";
        new MockUp<SecurityManagerFactory>() {
            @Mock
            public CustosSecurityManager getSecurityManager(){return mockedSecurityManager;};
        };
        new Expectations() {{ mockedSecurityManager.getUserInfoFromAuthzToken(authzToken); result = userInfo;}};

        UserProfile createdUserProfile = userProfileServiceHandler.initializeUserProfile(authzToken);
        assertNotNull(createdUserProfile);
        assertEquals(old_email_address,createdUserProfile.getEmails().get(0));
        assertEquals(userInfo.getUsername().toLowerCase(), createdUserProfile.getUserId());
        assertEquals(GATEWAY_ID,createdUserProfile.getGatewayId());
    }
    @Test
    public void testGetAllUserProfilesInGateway() throws UserProfileServiceException{
        List<UserProfile> allGatewayUsers = userProfileServiceHandler.getAllUserProfilesInGateway(authzToken, GATEWAY_ID, 0, 0);
        assertEquals(NUMBER_OF_USERS, allGatewayUsers.size());
    }
    @Test
    public void testDoesUserExistWhenUserExists()throws UserProfileServiceException{
        boolean shouldBeTrue = userProfileServiceHandler.doesUserExist(authzToken, "test-username", "test-gateway");
        assertTrue(shouldBeTrue);
    }
    @Test
    public void testDoesUserExistWhenUserDoesNotExist()throws UserProfileServiceException{
        boolean shouldBeFalse = userProfileServiceHandler.doesUserExist(authzToken, "does-not-exist", "default");
        assertFalse(shouldBeFalse);

    }
    @Test
    //TODO: check
    public void testDeleteUserProfile() throws UserProfileServiceException{
        UserProfile userProfile = userProfileServiceHandler.deleteUserProfile(authzToken, NEW_USER_ID, GATEWAY_ID);
        assertNotNull(userProfile);
        assertEquals(NEW_USER_ID, userProfile.getUserId());
        NUMBER_OF_USERS = 0;
        //testGetAllUserProfilesInGateway();
    }
}
