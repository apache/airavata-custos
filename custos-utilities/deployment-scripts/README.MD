### Custos Services Deployment Scripts

#### Prerequisites

- You should have a Kubernetes Cluster with at least one mater and three worker nodes. (Rancher or Kubespray can be 
  used to bootstrap K8 Cluster)
  
#### Rancher Installation Steps (Optional)
 
  - Install Rancher 2.5.5 or above on one of the VMs
  -  Spin Up a VM on JS2
  -  use https://github.com/CloudVE/cloudman-boot to boostrap a K8 cluster for rancher
  -  Login to that node
  -  helm repo add rancher https://releases.rancher.com/server-charts/stable
  -  kubectl create namespace cattle-system
  -  helm repo update
  -  helm install -n cattle-system rancher rancher/rancher --set hostname=HOSTNAME 
     --set ingress.tls.source=letsEncrypt --set letsEncrypt.email="admin@cloudve.org" 
     --set letsEncrypt.environment="production" --set letsEncrypt.ingress.class=nginx  
     --version=2.5.5
     

#### Install Services

- Install dependencies in the following order. Each folder contains installation stpes
- Login to the cluster master node.
- Install helm3  (/https://helm.sh/docs/intro/install/)

   - Cert-Manager
   - Keycloak
   - Vault
   - Custos Services 
 