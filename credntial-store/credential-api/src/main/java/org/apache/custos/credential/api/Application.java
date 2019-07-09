package org.apache.custos.credential.api;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@ComponentScan(basePackages = {"org.apache.custos", "org.apache.airavata"})
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}