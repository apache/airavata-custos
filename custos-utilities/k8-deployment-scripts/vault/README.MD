### Deploy HashiCorp Consul
Consul is the storage for vault.
####  Create PVs

Create three PVs for each mount point   /hashicorp/consul
 -  `kubectl apply -f pv.yaml`   

Create storage
- `kubectl apply -f storage.yaml`

Deploy consule
- `helm install consul hashicorp/consul --version 0.31.1 -n vault`


#### Deploy Vault

`helm install vault hashicorp/vault     --namespace vault     -f vaules.yaml     --version 0.10.0`

#### Deploy Ingress

`kubectl apply -f ingress.yaml -n vault`


    

 