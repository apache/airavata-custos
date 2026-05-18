"use client";

import { useState } from "react";
import type { Certificate } from "./types";
import { useRevokeCertificate } from "./hooks";

type Props = {
  cert: Certificate;
  onRevoked?: () => Promise<void> | void;
  triggerClassName?: string;
  confirmClassName?: string;
  cancelClassName?: string;
};

export function RevokeDialog({
  cert,
  onRevoked,
  triggerClassName = "rounded-md border px-3 py-2 text-sm",
  confirmClassName = "rounded-md bg-red-600 px-3 py-2 text-sm text-white",
  cancelClassName = "rounded-md border px-3 py-2 text-sm",
}: Props) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("User requested revocation");
  const [successMessage, setSuccessMessage] = useState("");
  const { revoke, loading, error } = useRevokeCertificate();

  async function onConfirm() {
    try {
      const response = await revoke({
        serial_number: cert.serial_number,
        reason,
      });

      setSuccessMessage(response.message || "Certificate revoked.");
      await onRevoked?.();
      setOpen(false);
    } catch {
      // The hook owns the rendered error state.
    }
  }

  if (!open) {
    return (
      <>
        <button
          className={triggerClassName}
          onClick={() => setOpen(true)}
          type="button"
        >
          Revoke
        </button>
      {successMessage && (
        <p className="mt-2 text-sm text-green-700">{successMessage}</p>
      )}
      </>
    );
  }

  return (
    <div className="absolute bottom-28 right-16 z-50 w-[360px] rounded-lg border border-neutral-200 bg-white p-5 text-left shadow-2xl">
      <h2 className="font-semibold">Revoke certificate?</h2>
      <p className="mt-1 text-sm text-neutral-500">
        This will revoke certificate serial {cert.serial_number}.
      </p>

      <label className="mt-3 block text-sm">
        Reason
        <input
          className="mt-1 w-full rounded-md border px-3 py-2"
          value={reason}
          onChange={(event) => setReason(event.target.value)}
        />
      </label>

      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}

      <div className="mt-4 flex gap-2">
        <button
          className={cancelClassName}
          onClick={() => setOpen(false)}
          type="button"
        >
          Cancel
        </button>
        <button
          className={confirmClassName}
          onClick={onConfirm}
          disabled={loading || !reason}
          type="button"
        >
          {loading ? "Revoking..." : "Confirm revoke"}
        </button>
      </div>
    </div>
  );
}
