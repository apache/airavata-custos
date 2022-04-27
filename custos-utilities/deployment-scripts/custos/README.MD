### Deploy Mysql
 Database server need to be installed.
####  Create PVs

Create three PVs for each mount point   /bitnami/mysql
 -  `kubectl apply -f pv.yaml`   

Then deploy mysql 

- `helm install mysql  bitnami/mysql -f values.yaml -n custos --version 8.8.8`

### Deploy Custos services
 
- Checkout Custos develop branch
-   `mvn clean install -P container`
-   `mvn antrun: run -P scp-to-remote`

You need to change  the remote  host and relevant parameters as necessary.



 