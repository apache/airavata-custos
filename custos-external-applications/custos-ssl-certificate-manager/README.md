# custos-ssl-certificate-manager

A ssl certificate manager using custos and acme4j. The purpose of this application is to run as a scheduled job and 
update the ssl certificates of provided domain.

![Architecture](https://i.ibb.co/xswjXnG/ssl-certificate-manager-1.png)

There are two sets of key pairs. One required for creating and accessing your account in Let's Encrypt, the other is 
required for encrypting the traffic on your domain. Assume these keys are user.key and domain.key. You can provide 
these as inputs or certificate manager will create keys for you. Following is a detailed explanation of the steps 1-7 
illustrated in architecture diagram above.

1. Cert manager stores user.key in custos for future use.
2. Order is placed in Let's Encrypt for certificate renewal.
3. Order provides HTTP-01 challenge details to certificate manager and certificate manager creates challenge file in 
   cert manager nginx server
4. Let's encrypt validates the challenge
5. On successful completion of 4 certificate manger creates generate a CSR for all of the domains, and sign it with the 
   domain.key and order the certificate
6. Cert manager receives the certificate and check whether a certificate already exists for the domain in custos. 
   If not the certificate is stored in custos along with the token received from the operation. If certificate 
   already exist cert manager updates the existing certificate without creating a new entry. Also domain key pair 
   in 5 is stored in custos for future reuse.
   
## Steps to run the certificate manager

### Setting up the challenge server

As described above we need to setup a challenge server to handle Let's encrypt http-01 challenge. As per the current 
implementation certificate manager is designed to use a nginx server that has [lua-nginx-module](https://github.com/openresty/stream-lua-nginx-module). Alternatively you can use [nginx-lua docker](https://github.com/fabiocicerchia/nginx-lua) image.
In order to certificate-manager to successfully communicate with the challenge server the default conf of nginx server 
in `/etc/nginx/conf.d/default.conf` should have following location directives along with your configurations.

```conf
server {
    location /ca/challenge {
        if ($request_method = POST ) {
            content_by_lua_block {
                local output = "echo -n ".. ngx.var.arg_content.." >> /var/www/letsencrypt/.well-known/acme-challenge/".. ngx.var.arg_file
                os.execute(output)
            }
        }

        if ($request_method = DELETE ) {
            content_by_lua_block {
                local output = "rm /var/www/letsencrypt/.well-known/acme-challenge/".. ngx.var.arg_file
                os.execute(output)
            }
        }
    }

    location ^~ /.well-known/acme-challenge/ {
        allow all;
        root /var/www/letsencrypt/;
        default_type text/plain;
    }
}
```

Challenge file will be created inside `/var/www/letsencrypt/.well-known/acme-challenge` and you need to create this 
folder manually. Following is an example of dockerfile that describes above steps. 

```Dockerfile
FROM fabiocicerchia/nginx-lua:alpine

COPY default.conf /etc/nginx/conf.d/default.conf

RUN mkdir -p /var/www/letsencrypt/.well-known/acme-challenge/
RUN chmod -R 777 /var/www/letsencrypt/.well-known/acme-challenge/
```

### Run certificate manager

1. Build custos-ssl-certificate-manager jar

`mvn clean install`

2. If you are running the plain jar you need to provide a location of a properties file as a CLI argument. The 
   content of the properties file will be as follows.

```properties
CRON_EXPRESSION=<cron_expression>
NGINX_URL=<challenge_server_url>
CA_URL=<CA_server_url>
CA_DOMAINS=<space_seperated_domains>
CA_USER_KEY_PATH=<certificate_authority_account_key_pair>
CA_DOMAIN_KEY_PATH=<private_key_for_encrypting_traffic>
CUSTOS_URL=<custos_url>
CUSTOS_PORT=<custos_port>
CUSTOS_CLIENT_ID=<custos_client_id>
CUSTOS_CLIENT_SECRET=<custos_client_secret>
CUSTOS_OWNER_ID=<custos_owner_id>
```
- `CRON_EXPRESSION`: Certificate manager uses quartz as job scheduling library. You must provide a valid cron 
  expression in order to successfully run the application. Eg- `0 0/1 * 1/1 * ? *` Runs in one minute intervals
- `CA_URL`: For testing purposed you can use `acme://letsencrypt.org/staging` and `acme://letsencrypt.org`for 
  production.
- `CA_DOMAINS`: Space separated domains you need to renew certificates for. Eg: `example.com example.org`

3. You can build a docker image using the jar. In this case properties should be provided as environment variables.