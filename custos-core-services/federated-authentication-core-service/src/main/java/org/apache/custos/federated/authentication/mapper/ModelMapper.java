/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.federated.authentication.mapper;

import org.apache.custos.federated.authentication.persistance.model.CILogonInstitution;
import org.apache.custos.federated.authentication.service.Institution;

public class ModelMapper {

    public static CILogonInstitution convert(long tenantId, String institutionId, String type, String performedBy) {


        String id = institutionId + "@" + tenantId;

        CILogonInstitution ciLogonInstitution = new CILogonInstitution();
        ciLogonInstitution.setId(id);
        ciLogonInstitution.setInstitutionId(institutionId);
        ciLogonInstitution.setTenantId(tenantId);
        ciLogonInstitution.setType(type);
        if (performedBy != null && !performedBy.equals("")) {
            ciLogonInstitution.setCreatedBy(performedBy);
        }

        return ciLogonInstitution;

    }


    public static Institution convert(CILogonInstitution cache) {

        Institution.Builder msg = Institution.newBuilder();

        msg.setEntityId(cache.getInstitutionId());


        return msg.build();
    }


    public static Institution convert(org.apache.custos.federated.services.clients.cilogon.CILogonInstitution ciLogonInstitution) {

        Institution.Builder msg = Institution.newBuilder();

        msg.setEntityId(ciLogonInstitution.getEntityId());
        msg.setDisplayName(ciLogonInstitution.getDisplayName());
        msg.setOrganizationName(ciLogonInstitution.getOrganizationName());
        msg.setRandS(ciLogonInstitution.isRandS());
        return msg.build();

    }
}
