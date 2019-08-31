package org.apache.custos.sharing.service.api;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.apache.custos.sharing.service.core.db.utils.SharingRegistryDBInitConfig;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class AppConfig {
    @Autowired
    SharingRegistryDBInitConfig sharingRegistryDBInitConfig;
    @Bean
    public SharingRegistryDBInitConfig SharingRegistryDBInitConfigBeanMapper(){
        return new SharingRegistryDBInitConfig();
    }
    @Bean
    public SharingRegistryService SharingRegistryServiceBeanMapper() throws ApplicationSettingsException {
        return new SharingRegistryService(sharingRegistryDBInitConfig);
    }
}
