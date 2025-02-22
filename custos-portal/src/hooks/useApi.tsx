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

import { useState, useEffect } from "react";
import { useAuth } from "react-oidc-context";

export const useApi = (url: string, options?: RequestInit) => {
  const auth = useAuth();
  const [data, setData] = useState<any | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!url) return; // Avoid calling fetch if the URL is empty

    const controller = new AbortController();
    const signal = controller.signal;

    const fetchData = async () => {
      setIsPending(true);
      setError(""); // Reset error before fetching

      try {
        const response = await fetch(url, {
          ...options,
          headers: {
            ...options?.headers,
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
          signal, // Pass the signal to fetch
        });

        if (!response.ok) throw new Error(response.statusText);
        const json = await response.json();

        if (!signal.aborted) {
          setData(json);
        }
      } catch (err) {
        if (!signal.aborted) {
          setError(`Error: ${err.message} - Could not fetch data.`);
        }
      } finally {
        if (!signal.aborted) {
          setIsPending(false);
        }
      }
    };

    fetchData();

    return () => controller.abort(); // Cleanup on unmount or dependency change
  }, [auth.user?.access_token, options, url]);

  return { data, isPending, error };
};
