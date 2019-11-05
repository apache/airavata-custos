import mockit.*;

import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.model.security.UserInfo;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.profile.model.user.UserProfile;
import org.apache.custos.profile.user.cpi.exception.UserProfileServiceException;
import org.apache.custos.profile.user.handler.UserProfileServiceHandler;
import org.apache.custos.security.manager.CustosSecurityManager;
import org.apache.custos.security.manager.SecurityManagerFactory;
import org.junit.Before;
import org.junit.Test;


import java.util.HashMap;
import java.util.List;

import static org.junit.Assert.*;


public class TestUserProfileServiceHandler {

    private static String GATEWAY_ID = "test-gateway";
    private static String USER_NAME = "default-user";
    private static AuthzToken authzToken = new AuthzToken();
    private static String NEW_USER_ID = "test-username";
    private int NUMBER_OF_USERS = 1;
    @Tested
    private static UserProfileServiceHandler userProfileServiceHandler = new UserProfileServiceHandler();
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
        //create a user
        UserInfo userInfo = new UserInfo();
        userInfo.setUsername("test-username");
        userInfo.setEmailAddress("test@test.com");
        userInfo.setFirstName("test-first-name");
        userInfo.setLastName("test-last-name");
        //create a user with same username
        UserInfo userInfoCopy = new UserInfo();
        userInfoCopy.setUsername("test-username");
        userInfoCopy.setEmailAddress("some-other-test@test.com");
        userInfoCopy.setFirstName("someother-test-first-name");
        userInfoCopy.setLastName("someother-test-last-name");
        AuthzToken authzTokenCopy = new AuthzToken();
        HashMap<String,String> map = new HashMap<>();
        map.put(Constants.GATEWAY_ID, GATEWAY_ID);
        map.put(Constants.USER_NAME, userInfoCopy.getUsername());
        authzTokenCopy.setClaimsMap(map);
        authzTokenCopy.setAccessToken("access token");

        new MockUp<SecurityManagerFactory>() {
            @Mock
            public CustosSecurityManager getSecurityManager(){return mockedSecurityManager;};
        };
        new Expectations() {{
            mockedSecurityManager.getUserInfoFromAuthzToken(authzToken); result = userInfo;
            mockedSecurityManager.getUserInfoFromAuthzToken(authzTokenCopy); result = userInfoCopy;}};

        userProfileServiceHandler.initializeUserProfile(authzToken);
        UserProfile createdUserProfile = userProfileServiceHandler.initializeUserProfile(authzTokenCopy);
                assertNotNull(createdUserProfile);
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
