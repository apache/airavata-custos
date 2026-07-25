FROM debian:bookworm-slim

RUN set -eu; \
    { \
      echo "slapd slapd/password1 password admin"; \
      echo "slapd slapd/password2 password admin"; \
      echo "slapd slapd/domain string custos.example"; \
      echo "slapd shared/organization string Custos"; \
    } | debconf-set-selections; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends slapd ldap-utils; \
    rm -rf /var/lib/apt/lists/*

COPY voperson-schema.ldif seed.ldif /tmp/

# Data is baked at build time; the container is a throwaway dev directory.
RUN set -eu; \
    slapadd -F /etc/ldap/slapd.d -n 0 -l /tmp/voperson-schema.ldif; \
    slapadd -F /etc/ldap/slapd.d -n 1 -l /tmp/seed.ldif; \
    chown -R openldap:openldap /etc/ldap/slapd.d /var/lib/ldap

EXPOSE 389

CMD ["slapd", "-h", "ldap:///", "-u", "openldap", "-g", "openldap", "-d", "256"]
