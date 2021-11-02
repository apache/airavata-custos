FROM quay.io/keycloak/keycloak:9.0.2
COPY src/main/resources/standalone-ha.xml /opt/jboss/keycloak/standalone/configuration/standalone-ha.xml
COPY src/main/resources/standalone.xml /opt/jboss/keycloak/standalone/configuration/standalone.xml
COPY src/main/resources/themes  /opt/jboss/keycloak/themes/