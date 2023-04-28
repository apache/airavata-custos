FROM nginx:stable-alpine
COPY src/main/containers/nginx.conf /etc/nginx/conf.d/default.conf
COPY src/main/containers/certificate_mul.pem /etc/nginx/certificate.pem
COPY src/main/containers/key_mul.pem /etc/nginx/key.pem
EXPOSE 443