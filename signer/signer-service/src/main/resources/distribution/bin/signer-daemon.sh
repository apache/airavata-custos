#!/bin/bash
#
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
#

# Custos Signer Service startup script

JAVA_HOME=${JAVA_HOME:-/usr/lib/jvm/java-17-openjdk}
JAVA_OPTS=${JAVA_OPTS:-"-Xmx2g -Xms1g"}
SIGNER_HOME=${SIGNER_HOME:-$(dirname "$0")/..}
SIGNER_CONF=${SIGNER_CONF:-$SIGNER_HOME/conf}
SIGNER_LOG=${SIGNER_LOG:-$SIGNER_HOME/logs}

mkdir -p "$SIGNER_LOG"

CLASSPATH="$SIGNER_HOME/lib/*"

JVM_OPTS="$JAVA_OPTS"
JVM_OPTS="$JVM_OPTS -Dspring.config.location=classpath:/application.yml,file:$SIGNER_CONF/application.yml"
JVM_OPTS="$JVM_OPTS -Dlogging.config=classpath:logback-spring.xml"
JVM_OPTS="$JVM_OPTS -Dspring.profiles.active=production"

JVM_OPTS="$JVM_OPTS -Dcustos.signer.home=$SIGNER_HOME"
JVM_OPTS="$JVM_OPTS -Dcustos.signer.conf=$SIGNER_CONF"
JVM_OPTS="$JVM_OPTS -Dcustos.signer.log=$SIGNER_LOG"

echo "Starting Custos Signer Service..."
echo "JAVA_HOME: $JAVA_HOME"
echo "SIGNER_HOME: $SIGNER_HOME"
echo "SIGNER_CONF: $SIGNER_CONF"
echo "SIGNER_LOG: $SIGNER_LOG"
echo "JVM_OPTS: $JVM_OPTS"

exec "$JAVA_HOME/bin/java" $JVM_OPTS -cp "$CLASSPATH" org.apache.custos.signer.service.SignerApplication "$@"
