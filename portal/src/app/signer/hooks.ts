// React hooks that wrap the signer API. They expose a small {data, loading,
// error} shape per call site so client components can render skeletons,
// error banners, and live data without each owning their own request state.
import { useCallback, useEffect, useState } from "react";
import {
  getCertificate,
  getUserInfo,
  listCertificates,
  revokeCertificate,
} from "./api";
import type {
  Certificate,
  CertificateListResponse,
  RevokeRequest,
  UserInfo,
} from "./types";

type AsyncState<T> = {
  data: T | null;
  loading: boolean;
  error: string | null;
};

function initialState<T>(): AsyncState<T> {
  return {
    data: null,
    loading: true,
    error: null,
  };
}

// These lightweight hooks keep data fetching local to the client components
// that need live signer state.
export function useUserInfo() {
  const [state, setState] = useState<AsyncState<UserInfo>>(initialState);

  useEffect(() => {
    getUserInfo()
      .then((data) => setState({ data, loading: false, error: null }))
      .catch((error: Error) =>
        setState({ data: null, loading: false, error: error.message })
      );
  }, []);

  return state;
}

export function useCertificates(limit = 20, offset = 0, username = "") {
  const [state, setState] =
    useState<AsyncState<CertificateListResponse>>(initialState);

  const reload = useCallback(() => {
    // Preserve the last successful data while a manual refresh is in flight.
    setState((prev) => ({ ...prev, loading: true, error: null }));

    listCertificates({
      limit,
      offset,
      username: username.trim() || undefined,
    })
      .then((data) => setState({ data, loading: false, error: null }))
      .catch((error: Error) =>
        setState({ data: null, loading: false, error: error.message })
      );
  }, [limit, offset, username]);

  useEffect(() => {
    reload();
  }, [reload]);

  return { ...state, reload };
}

export function useCertificate(serial?: string) {
  const [state, setState] = useState<AsyncState<Certificate>>(initialState);

  const reload = useCallback(() => {
    if (!serial) {
      setState({ data: null, loading: false, error: null });
      return;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));

    getCertificate(serial)
      .then((data) => setState({ data, loading: false, error: null }))
      .catch((error: Error) =>
        setState({ data: null, loading: false, error: error.message })
      );
  }, [serial]);

  useEffect(() => {
    reload();
  }, [reload]);

  return { ...state, reload };
}

export function useRevokeCertificate() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function revoke(payload: RevokeRequest) {
    setLoading(true);
    setError(null);

    try {
      return await revokeCertificate(payload);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Revoke failed";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }

  return { revoke, loading, error };
}
