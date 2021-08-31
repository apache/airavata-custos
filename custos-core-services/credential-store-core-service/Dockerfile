FROM openjdk:11.0.5-jdk-slim
COPY src/main/resources/vault-client-truststore.pkcs12 /home/ubuntu/vault-client-truststore.pkcs12
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]