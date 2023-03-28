FROM nginx:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY certificate.pem /etc/nginx/certificate.pem
COPY key.pem /etc/nginx/key.pem
EXPOSE 443