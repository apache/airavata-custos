/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {
  BreadcrumbLabelsProvider,
  useBreadcrumbLabels,
} from "@/shared/layout/BreadcrumbLabelsProvider";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const ID = "29377f4d-0000-4000-8000-000000000001";
vi.mock("next/navigation", () => ({ usePathname: () => `/admin/amie/packets/${ID}` }));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));

import { AmiePage } from "../AmieNav";

function Crumb() {
  return <p data-testid="crumb">{useBreadcrumbLabels()[ID]}</p>;
}

describe("<AmiePage />", () => {
  // Make sure a caller without packet read sees "Packet" in the crumb, never the packet's id.
  it("labels a gated packet crumb", () => {
    render(
      <BreadcrumbLabelsProvider>
        <AmiePage description="AMIE packet" subject="AmiePacket">
          <p>packet</p>
        </AmiePage>
        <Crumb />
      </BreadcrumbLabelsProvider>,
    );
    expect(screen.getByTestId("crumb")).toHaveTextContent("Packet");
  });
});
