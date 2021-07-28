FROM openjdk:11.0.5-jdk-slim
COPY src/main/resources/keycloak-client-truststore.pkcs12 /home/ubuntu/keystore/keycloak-client-truststore.pkcs12
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Djavax.net.debug=ssl:handshake:verbose:keymanager:trustmanager -Djava.security.debug=access:stack -Dhttps.protocols=TLSv1,TLSv1.1,TLSv1.2","-jar","/app.jar"]