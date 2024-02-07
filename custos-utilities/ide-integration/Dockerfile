FROM nginx:stable-alpine
COPY src/main/containers/nginx.conf /etc/nginx/conf.d/default.conf.template
COPY src/main/containers/certificate_mul.pem /etc/nginx/certificate.pem
COPY src/main/containers/key_mul.pem /etc/nginx/key.pem
COPY src/main/containers/init/nginx/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
EXPOSE 443