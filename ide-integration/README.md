## Introduction 

This module is to setup custos authentication, profile services inside Intelij IDEA for development purposes

## Prerequisites

* Docker installed with 'docker-compose' utility
  https://docs.docker.com/compose/

* InteliJ IDEA with Java 8 installed
  https://www.jetbrains.com/idea/download/

* Maven

* Git

## Steps

### Setting up the development environment

* Clone Custos repository to a local directory

  ```
  git clone https://github.com/apache/airavata-custos
  ```

* Checkout develop branch
  ```
  git checkout develop
  ```
* Open the project using InteliJ IDEA
  
* Browse to ide-integration -> custos-services -> src-> main -> resources

* start the docker containers 

    ```
    docker-compose up -d
    ```
* check if docker containers are up
    ```
    docker-compose ps
    ``` 

* Build the develop branch using Maven


### Starting the services

* Go to org.apache.custos.server.start.CustosAPIServerStarter class and right click on the editor and click Run option. This will start profile and authentication server.
* The port on which profile and authentication service run can be changed in the custos-server.properties file found in the resources folder.
* Once the services are up and running. Go to org.apache.custos.profile.service.samples to run code samples for tenant and profile services and org.apache.custos.authentication.service.sample.CustosAuthenticationServiceSample

### Stop all components

* For each composer file, run following commands to cleanup docker spawned components

  ```
  docker-compose down
  ```
 
  ```
  docker-compose rm
  ```
