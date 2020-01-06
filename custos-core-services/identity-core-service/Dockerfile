FROM openjdk:11-jre-slim
COPY src/main/resources/keycloak-client-truststore.pkcs12 /home/ubuntu/keystore/keycloak-client-truststore.pkcs12
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]