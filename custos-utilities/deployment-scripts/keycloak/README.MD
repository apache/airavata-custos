### Deploy Postgresql
 Database server need to be installed.
####  Create PVs

Create three PVs for each mount point   /bitnami/postgresql
 -  `kubectl apply -f pv.yaml`   

Then deploy postgresql 

- `helm install keycloak-db-postgresql bitnami/postgresql -f postgresql-values.yaml -n keycloak --version 10.12.3`

### Deploy Keycloak
 
Deploy Keycloak

  - Deploy OLM
    - https://operatorhub.io/how-to-install-an-operator
    - https://github.com/operator-framework/operator-lifecycle-manager/issues/854
      
  - Deploy keycloak operator 
    -  Download the keycloak-operator and copy operator.yaml file to deploy/ folder
    - make cluster/prepare
    - `kubectl apply -f deploy/operator.yaml -n keycloak`
     
    You might need to edit roles to add apiGroup networking.k8s.io and give permissions to ingresses 
    -  `kubectl edit roles keycloak-operator  -n keycloak ` 
        
  - Deploy keycloak
    `kubectl apply -f keycloak-db-secret.yaml -n keycloak`
    `kubectl apply -f custos-keycloak.yaml -n keycloak`
    
  - Deploy ingress controller in ingress-nginx namespace
    `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.44.0/deploy/static/provider/baremetal/deploy.yaml`


    

 