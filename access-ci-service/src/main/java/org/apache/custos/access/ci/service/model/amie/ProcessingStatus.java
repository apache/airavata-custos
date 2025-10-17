/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */
package org.apache.custos.access.ci.service.model.amie;

public enum ProcessingStatus {

    /**
     * The event has been created and is waiting to be processed by a worker.
     */
    NEW,
    /**
     * A worker has picked up the event and is actively processing it.
     */
    RUNNING,
    /**
     * The event was processed successfully.
     */
    SUCCEEDED,
    /**
     * The event failed processing and will not be automatically retried.
     */
    FAILED,
    /**
     * The event failed a previous attempt and is waiting to be retried.
     */
    RETRY_SCHEDULED
}
