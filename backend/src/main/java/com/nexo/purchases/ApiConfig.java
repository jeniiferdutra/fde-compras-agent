package com.nexo.purchases;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.support.TransactionTemplate;
import org.springframework.web.client.RestClient;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class ApiConfig {
    @Bean TransactionTemplate transactionTemplate(PlatformTransactionManager manager) { return new TransactionTemplate(manager); }

    @Bean WebMvcConfigurer corsConfigurer() {
        return new WebMvcConfigurer() {
            @Override public void addCorsMappings(CorsRegistry registry) {
                registry.addMapping("/**").allowedOrigins("*").allowedMethods("GET", "POST", "OPTIONS").allowedHeaders("*");
            }
        };
    }
}
