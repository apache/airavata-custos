import { useState, useEffect } from "react";
import { useAuth } from "react-oidc-context";

export const useApi = (url: string, options?: RequestInit) => {
  const auth = useAuth();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [data, setData] = useState<any | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    const fetchData = async () => {
      setIsPending(true);
      try {
        const response = await fetch(url, {
          ...options,
          headers: {
            ...options?.headers,
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        });
        if (!response.ok) throw new Error(response.statusText);
        const json = await response.json();
        setIsPending(false);
        setData(json);
        setError("");
      } catch (error) {
        setError(`${error} Could not Fetch Data.`);
        setIsPending(false);
      }
    };
    fetchData();
  }, [auth.user?.access_token, options, url]);
  return { data, isPending, error };
};
