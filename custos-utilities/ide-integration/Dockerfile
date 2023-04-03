FROM nginx:stable-alpine
COPY src/main/containers/nginx.conf /etc/nginx/conf.d/default.conf
COPY src/main/containers/certificate.pem /etc/nginx/certificate.pem
COPY src/main/containers/key.pem /etc/nginx/key.pem
EXPOSE 443