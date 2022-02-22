# Apache Custos Jupyterhub Authenticator
The Apache Custos Jupyterhub Authenticator provides CILogon based federated authentication for Jupyterhub. In addtion it provides access to Custos IAM solutions such as
 - Fine-Grained Authorization

 - Secret Management

- Service Accounts ..etc.

Additional Information : https://airavata.apache.org/custos/


### Folder Structure

 - custosauthenticator
      
    Includes oidc plugin to connect with custos oauth services 
    
 ### Configuration
 Add following configuration to Jupyter  Hub in helm chart configuration (values.yaml). You need to request an tenant from 
 Custos Portal to start
 
 - dev:  https://dev.portal.usecustos.org/
 - production: https://portal.usecustos.org/
 
 #### Hub configuration
 ```
 hub:
  config:
    CustosOAuthenticator:
      client_id: CHANGE_ME
      client_secret: CHANGE_ME
      oauth_callback_url: https://<jupyter_host>/hub/oauth_callback
      custos_host: custos.scigap.org      
    JupyterHub:     
      authenticator_class: custosauthenticator.custos.CustosOAuthenticator 
 ```
    
    
    
              
 
              


 
  