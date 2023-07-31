FROM openjdk:17.0.2-slim
COPY src/main/resources/certificate_mul.p12  /home/ubuntu/keystore/keycloak-client-truststore.pkcs12
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Xms130m", "-Xmx130m", "-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]