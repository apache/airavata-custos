FROM rockylinux:9

RUN dnf install -y openssh-server openssh-clients && dnf clean all && \
    useradd -m testuser && useradd -m testuser2 && \
    mkdir -p /etc/ssh/auth_principals && \
    echo testuser > /etc/ssh/auth_principals/testuser && \
    echo testuser2 > /etc/ssh/auth_principals/testuser2

COPY 50-cert-auth.conf /etc/ssh/sshd_config.d/50-cert-auth.conf
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
