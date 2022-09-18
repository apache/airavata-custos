FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common \
    python3-pip

RUN pip3 install pytest pytest-testinfra ansible && mkdir -p /conjurinc/

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"
RUN apt-get update && apt-get -y install docker-ce
RUN apt-get update && apt-get install -y gcc build-essential
RUN apt-add-repository -y ppa:brightbox/ruby-ng && apt-get update && apt-get install -y ruby2.4 ruby2.4-dev
RUN gem install conjur-cli

WORKDIR /conjurinc/

CMD ["/bin/sleep", "1d"]
