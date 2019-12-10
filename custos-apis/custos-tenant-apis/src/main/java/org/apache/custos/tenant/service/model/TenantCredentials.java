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
package org.apache.custos.tenant.service.model;

import java.util.Objects;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonCreator;
import org.springframework.validation.annotation.Validated;
import javax.validation.Valid;
import javax.validation.constraints.*;

/**
 * TenantCredentials
 */
@Validated

public class TenantCredentials   {
  @JsonProperty("custosClientId")
  private String custosClientId = null;

  @JsonProperty("custosClientSecret")
  private String custosClientSecret = null;

  @JsonProperty("iamClientId")
  private String iamClientId = null;

  @JsonProperty("iamClientSecret")
  private String iamClientSecret = null;

  public TenantCredentials custosClientId(String custosClientId) {
    this.custosClientId = custosClientId;
    return this;
  }

  /**
   * Get custosClientId
   * @return custosClientId
  **/

  @NotNull


  public String getCustosClientId() {
    return custosClientId;
  }

  public void setCustosClientId(String custosClientId) {
    this.custosClientId = custosClientId;
  }

  public TenantCredentials custosClientSecret(String custosClientSecret) {
    this.custosClientSecret = custosClientSecret;
    return this;
  }

  /**
   * Get custosClientSecret
   * @return custosClientSecret
  **/

  @NotNull


  public String getCustosClientSecret() {
    return custosClientSecret;
  }

  public void setCustosClientSecret(String custosClientSecret) {
    this.custosClientSecret = custosClientSecret;
  }

  public TenantCredentials iamClientId(String iamClientId) {
    this.iamClientId = iamClientId;
    return this;
  }

  /**
   * Get iamClientId
   * @return iamClientId
  **/

  @NotNull


  public String getIamClientId() {
    return iamClientId;
  }

  public void setIamClientId(String iamClientId) {
    this.iamClientId = iamClientId;
  }

  public TenantCredentials iamClientSecret(String iamClientSecret) {
    this.iamClientSecret = iamClientSecret;
    return this;
  }

  /**
   * Get iamClientSecret
   * @return iamClientSecret
  **/

  @NotNull


  public String getIamClientSecret() {
    return iamClientSecret;
  }

  public void setIamClientSecret(String iamClientSecret) {
    this.iamClientSecret = iamClientSecret;
  }


  @Override
  public boolean equals(java.lang.Object o) {
    if (this == o) {
      return true;
    }
    if (o == null || getClass() != o.getClass()) {
      return false;
    }
    TenantCredentials tenantCredentials = (TenantCredentials) o;
    return Objects.equals(this.custosClientId, tenantCredentials.custosClientId) &&
        Objects.equals(this.custosClientSecret, tenantCredentials.custosClientSecret) &&
        Objects.equals(this.iamClientId, tenantCredentials.iamClientId) &&
        Objects.equals(this.iamClientSecret, tenantCredentials.iamClientSecret);
  }

  @Override
  public int hashCode() {
    return Objects.hash(custosClientId, custosClientSecret, iamClientId, iamClientSecret);
  }

  @Override
  public String toString() {
    StringBuilder sb = new StringBuilder();
    sb.append("class TenantCredentials {\n");
    
    sb.append("    custosClientId: ").append(toIndentedString(custosClientId)).append("\n");
    sb.append("    custosClientSecret: ").append(toIndentedString(custosClientSecret)).append("\n");
    sb.append("    iamClientId: ").append(toIndentedString(iamClientId)).append("\n");
    sb.append("    iamClientSecret: ").append(toIndentedString(iamClientSecret)).append("\n");
    sb.append("}");
    return sb.toString();
  }

  /**
   * Convert the given object to string with each line indented by 4 spaces
   * (except the first line).
   */
  private String toIndentedString(java.lang.Object o) {
    if (o == null) {
      return "null";
    }
    return o.toString().replace("\n", "\n    ");
  }
}

