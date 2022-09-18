FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /cyberark

# install ansible
RUN apt-get update && \
    apt-get install -y ansible

# install python 3
RUN apt-get update && \
    apt-get install -y python3-pip && \
    pip3 install --upgrade pip==9.0.3

# install ansible and its test tool
RUN pip3 install ansible pytest-testinfra

# install docker installation requirements
RUN apt-get update && \
    apt-get install -y apt-transport-https \
                       ca-certificates \
                       curl \
                       software-properties-common

# install docker
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"

RUN apt-get update && \
    apt-get -y install docker-ce
