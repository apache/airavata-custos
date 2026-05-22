FROM slurmrest/base:24.05

RUN echo "PermitRootLogin yes"            >> /etc/ssh/sshd_config \
 && echo "PasswordAuthentication yes"     >> /etc/ssh/sshd_config \
 && echo "root:rootpass" | chpasswd

COPY scripts/entrypoint-login.sh /usr/local/bin/entrypoint-login.sh
RUN chmod +x /usr/local/bin/entrypoint-login.sh

EXPOSE 22
ENTRYPOINT ["/usr/local/bin/entrypoint-login.sh"]
