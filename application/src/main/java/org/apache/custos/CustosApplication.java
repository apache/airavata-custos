package org.apache.custos;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;

@SpringBootApplication
@EnableJpaAuditing
public class CustosApplication {
    public static void main(String[] args) {
        SpringApplication.run(CustosApplication.class, args);
    }
}
