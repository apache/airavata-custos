# Airavata Custos Python SDK

Folder structure 

- airavata_custos : client
    - admin
    - security
    
- custos: thrift generated service APIs and models
- tests: test cases

Create a virtual environment

    python3 -m venv venv

Activate the virtual environment
    
    source venv/bin/activate

Install dependencies
    
    pip install -r requirements_dev.txt

Server configuration should be kept in a INI file in the following format. For more information refer to sample_settings.ini file  
    
    [IAMServerSettings]
    KEYCLOAK_AUTHORIZE_URL =
    KEYCLOAK_TOKEN_URL = 
    KEYCLOAK_USERINFO_URL = 
    KEYCLOAK_LOGOUT_URL = 
    VERIFY_SSL = 
    
    [ProfileServerSettings]
    PROFILE_SERVICE_HOST = 
    PROFILE_SERVICE_PORT =
    
Keycloak connections

    - authenticate_user
    - authenticate_account
    - authenticate_using_refresh_token

Admin operations
    
    - is_username_available
    - register_user
    - is_user_enabled
    - enable_user
    - delete_user
    - is_user_exist
    - get_user
    - get_users
    - reset_user_password