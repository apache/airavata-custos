<!--
    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.
-->

# Apache Airavata Custos Security

[![License](http://img.shields.io/badge/license-Apache--2-blue.svg?style=flat)](https://apache.org/licenses/LICENSE-2.0)
[![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/apache/airavata-custos)](https://github.com/apache/airavata-custos/pulls?q=is%3Apr+is%3Aclosed)
[![Build Status](https://travis-ci.org/apache/airavata-custos.png?branch=develop)](https://travis-ci.org/github/apache/airavata-custos)

Science gateways represent potential targets for cybersecurity threats to users, scientific research, and scientific resources. Custos is a software framework that provides common security operations for science gateways, including user identity and access management, gateway tenant profile management, resource secrets management, and groups and sharing management. The goals of the Custos project are to provide these services to a wide range of science gateway frameworks, providing the community with an open-source, transparent, and reviewed code base for common security operations; and to operate trustworthy security services for the science gateway community using this software base. To accomplish these goals, we implement Custos using a scalable microservice architecture that can provide highly available, fault-tolerant operations. Custos exposes these services through a language-independent Application Programming Interface that encapsulates science gateway usage scenarios.

**To find out more, please check out the [Custos website](https://airavata.apache.org/custos/) and the [Custos wiki](https://cwiki.apache.org/confluence/display/CUSTOS/Home).**

##Custos POM changes for MAC M1:


For successfully running the maven build for custom on Mac M1 laptops following changes are needed in the parent configuration files . These changes are related to protobuf and gen-grpc-java plugins as the the compatible version with Mac M1 are needed for both . These changes should be made at both the places where the plugins are specified .The architecture of the Mac M1 which is osx-x86_64 has to be passed along with versions for build to be successful.


    <configuration>
        <protocArtifact>com.google.protobuf:protoc:3.1.0:exe:osx-x86_64
        </protocArtifact>
        <pluginId>grpc-java</pluginId>
        <pluginArtifact>io.grpc:protoc-gen-grpc-java:1.0.0:exe:osx-x86_64
        </pluginArtifact>
    </configuration>




## Quickstart

## Installation Instructions

## Roadmap

## Contributing

## Questions or need help?

