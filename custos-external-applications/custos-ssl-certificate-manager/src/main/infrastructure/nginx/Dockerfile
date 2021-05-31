FROM fabiocicerchia/nginx-lua:alpine

COPY default.conf /etc/nginx/conf.d/default.conf

RUN mkdir -p /var/www/letsencrypt/.well-known/acme-challenge/
RUN chmod -R 777 /var/www/letsencrypt/.well-known/acme-challenge/
