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

"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as React from "react";
import {
  generateCertificate,
  getCertificate,
  getMyCertificate,
  getUserEmail,
  listCertificates,
  listMyCertificates,
  revokeCertificate,
  revokeMyCertificate,
} from "./api";
import type { GenerateCertificateInput } from "./schemas";
import type { CertificateListParams, CertificateListPaging } from "./types";

export const signerKeys = {
  all: ["signer-certificates"] as const,
  list: (params: CertificateListParams = {}) => [...signerKeys.all, "list", params] as const,
  detail: (serial: number | string) => [...signerKeys.all, "detail", String(serial)] as const,
  myList: (params: CertificateListParams = {}) => [...signerKeys.all, "me", "list", params] as const,
  myDetail: (serial: number | string) => [...signerKeys.all, "me", "detail", String(serial)] as const,
};

export function useUserEmail(id: string | null): string | undefined {
  const query = useQuery({
    queryKey: ["user-email", id],
    queryFn: () => getUserEmail(id ?? ""),
    enabled: id !== null,
    retry: false,
    staleTime: 300_000,
  });
  return query.data;
}

export function useCursorPaging() {
  // The cursor of each page after the first, so Previous can walk back.
  const [cursors, setCursors] = React.useState<string[]>([]);
  const [pageSize, setPageSize] = React.useState(20);
  const paging = (nextCursor: string | undefined): CertificateListPaging => ({
    page: cursors.length + 1,
    pageSize,
    hasNext: Boolean(nextCursor),
    onNext: () => {
      if (nextCursor) setCursors((prev) => [...prev, nextCursor]);
    },
    onPrevious: () => setCursors((prev) => prev.slice(0, -1)),
    onPageSizeChange: (next) => {
      setCursors([]);
      setPageSize(next);
    },
  });
  return { params: { limit: pageSize, cursor: cursors.at(-1) }, paging };
}

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useCertificates(params: CertificateListParams = {}) {
  return useQuery({
    queryKey: signerKeys.list(params),
    queryFn: () => listCertificates(params),
    ...DEFAULTS,
  });
}

export function useCertificate(serial: number | string | undefined) {
  return useQuery({
    queryKey:
      serial !== undefined && serial !== ""
        ? signerKeys.detail(serial)
        : [...signerKeys.all, "detail", "none"],
    queryFn: () => getCertificate(serial as number | string),
    enabled: serial !== undefined && serial !== "",
    ...DEFAULTS,
  });
}

export function useMyCertificates(params: CertificateListParams = {}) {
  return useQuery({
    queryKey: signerKeys.myList(params),
    queryFn: () => listMyCertificates(params),
    ...DEFAULTS,
  });
}

export function useMyCertificate(serial: string) {
  return useQuery({
    queryKey: signerKeys.myDetail(serial),
    queryFn: () => getMyCertificate(serial),
    ...DEFAULTS,
  });
}

export function useGenerateCertificate() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (input: GenerateCertificateInput) => generateCertificate(input),
    onSuccess: () => client.invalidateQueries({ queryKey: signerKeys.all }),
  });
}

export function useRevokeMyCertificate() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ serial, reason }: { serial: number; reason: string }) =>
      revokeMyCertificate(serial, reason),
    onSuccess: () => client.invalidateQueries({ queryKey: signerKeys.all }),
  });
}

export function useRevokeCertificate() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ serial, reason }: { serial: number; reason: string }) =>
      revokeCertificate(serial, reason),
    // Re-fetch from the backend rather than faking revoked state in React.
    onSuccess: (_data, { serial }) => {
      client.invalidateQueries({ queryKey: signerKeys.all });
      client.invalidateQueries({ queryKey: signerKeys.detail(serial) });
    },
  });
}
