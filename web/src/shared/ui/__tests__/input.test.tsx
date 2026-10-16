// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Field } from "../FormDialog";

describe("<Input required />", () => {
  it("refuses a value of only spaces, which the API would trim to nothing, in words", () => {
    render(<Field label="Name" name="name" required />);
    const input = screen.getByLabelText("Name") as HTMLInputElement;
    input.value = "   ";
    expect(input.checkValidity()).toBe(false);
    expect(input.validationMessage).toBe("Enter a name, not only spaces.");
    fireEvent.input(input, { target: { value: " gpu-01 " } });
    expect(input.checkValidity()).toBe(true);
  });
});
