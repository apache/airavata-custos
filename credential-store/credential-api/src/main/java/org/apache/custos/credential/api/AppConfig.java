package org.apache.custos.credential.api;

import org.dozer.DozerBeanMapper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class AppConfig {

    // Use for future custom mapping scenarios https://www.java-success.com/dozer-with-spring-maven-tutorial/
    @Bean
    public DozerBeanMapper dozerBeanMapper() {
        return new DozerBeanMapper();
    }
}
