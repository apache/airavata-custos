FROM jboss/keycloak:7.0.0
COPY src/main/resources/standalone-ha.xml /opt/jboss/keycloak/standalone/configuration/standalone-ha.xml
COPY src/main/resources/keycloak-metrics-spi-2.5.1.jar /opt/jboss/keycloak/standalone/deployments/
COPY src/main/resources/standalone.xml /opt/jboss/keycloak/standalone/configuration/standalone.xml
COPY src/main/resources/themes  /opt/jboss/keycloak/themes/