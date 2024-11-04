FROM quay.io/keycloak/keycloak:20.0.3 as builder

ENV KC_METRICS_ENABLED=true
ENV KC_HEALTH_ENABLED=true
ENV KC_FEATURES=preview
ENV KC_DB=postgres
ENV KC_HTTP_RELATIVE_PATH=/auth

COPY ./cache-ispn-jdbc-ping.xml /opt/keycloak/conf/cache-ispn-jdbc-ping.xml
ENV KC_CACHE_CONFIG_FILE=cache-ispn-jdbc-ping.xml

RUN /opt/keycloak/bin/kc.sh build

FROM quay.io/keycloak/keycloak:20.0.3

USER root
RUN microdnf update -y && \
    microdnf install -y jq && \
    microdnf clean all

USER keycloak
COPY --from=builder /opt/keycloak /opt/keycloak

WORKDIR /opt/keycloak

RUN keytool -genkeypair -storepass password -storetype PKCS12 -keyalg RSA -keysize 2048 -dname "CN=server" -alias server -ext "SAN:c=DNS:localhost,IP:127.0.0.1" -keystore conf/server.keystore

COPY docker-entrypoint.sh /docker-entrypoint.sh

ENV FORMAT_MESSAGES_PATTERN_DISABLE_LOOKUPS true

EXPOSE 7800
ENTRYPOINT ["/docker-entrypoint.sh"]
