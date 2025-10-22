# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

FROM openjdk:17-jre-slim

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r custos && useradd -r -g custos custos

RUN mkdir -p /opt/custos/signer-service \
    && mkdir -p /var/lib/custos/krl \
    && mkdir -p /var/log/custos \
    && chown -R custos:custos /opt/custos \
    && chown -R custos:custos /var/lib/custos \
    && chown -R custos:custos /var/log/custos

COPY target/custos-signer-service-*.jar /opt/custos/signer-service/app.jar

COPY src/main/resources/application.yml /opt/custos/signer-service/application.yml

WORKDIR /opt/custos/signer-service

USER custos

EXPOSE 8084 9095

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8084/actuator/health || exit 1

ENTRYPOINT ["java", "-jar", "app.jar"]
