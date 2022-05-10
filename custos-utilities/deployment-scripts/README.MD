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
 

##### Deploying Custos Services

- First change the following parameters in pom.xml
    - <ssh.username>CHANGE_ME</ssh.username>   ( local ssh key password)
    - <ssh.privatekey>CHANGE_ME</ssh.privatekey> (local ssh privatekey)
    -  <host> CHANGE_ME</host> ( K8 cluster master rhost)
    - <spring.profiles.active>dev</spring.profiles.active> (current active profile)
    - <vault.token>CHANGE_ME</vault.token> ( vault token)
    - <iam.dev.username>CHANGE_ME}</iam.dev.username> ( keycloak  username)
    - <iam.dev.password>CHANGE_ME}</iam.dev.password> ( keycloak password)
    - <cilogon.dev.id>CHANGE_ME</cilogon.dev.id> (cilogon client id, contact help@cilogon.org to obtain admin tenant) 
    - <cilogon.dev.sec>CHANGE_ME</cilogon.dev.sec> (cilogon client sec)
    - <custos.email.password>CHANGE_ME</custos.email.password> (your email)
    - <spring.datasource.username>CHANGE_ME</spring.datasource.username> (mysql database username)
    - <spring.datasource.password>CHANGE_ME</spring.datasource.password> (myslq password)
    - <docker.image.prefix>CHANGEME</docker.image.prefix> (dockerhub image prefix)
    - <docker.image.repo>CHAGEME</docker.image.repo> (dockerhub repo account)
    
-  Build code
    `mvn clean install -P container`

- Push code images to repo
   `mvn dockerfile:push -P container`

-  deploy artifacts
   `mvn antrun:run -P scp-to-remote`
