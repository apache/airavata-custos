package org.apache.custos.sharing.service.api.constants;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.apache.custos.commons.utils.ServerSettings;
import org.springframework.stereotype.Component;

@Component
public class SharingRegistryEndpoints {

    public String BASE_URL;

    //group APIs
    public String CREATE_GROUP = "/group";
    public String UPDATE_GROUP = "/group";
    public String IS_GROUP_EXISTS = "/group/exists/id/{groupId}/domain/{domainId}";
    public String DELETE_GROUP = "/group/id/{groupId}/domain/{domainId}";
    public String GET_GROUP = "/group/id/{groupId}/domain/{domainId}";

    SharingRegistryEndpoints() throws ApplicationSettingsException {
        BASE_URL = ServerSettings.getSharingRegistryServerHost() + ":" + ServerSettings.getSharingRegistryServerPort();
    }
}
