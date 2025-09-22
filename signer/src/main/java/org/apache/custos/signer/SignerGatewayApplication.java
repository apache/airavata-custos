package org.apache.custos.signer;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;

@SpringBootApplication
@EnableJpaAuditing
public class SignerGatewayApplication {
    public static void main(String[] args) {
        SpringApplication.run(SignerGatewayApplication.class, args);
    }
}
