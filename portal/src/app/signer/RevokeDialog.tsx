// Modal that confirms a certificate revocation and submits it through
// useRevokeCertificate. Lives next to the detail page since it's the only
// caller, but is split out so it can be unit-tested independently later.
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import type { Certificate } from "./types";
import { useRevokeCertificate } from "./hooks";

type Props = {
  cert: Certificate;
  onRevoked?: () => Promise<void> | void;
};

export function RevokeDialog({ cert, onRevoked }: Props) {
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

  return (
    <>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger
          render={<Button variant="destructive">Revoke</Button>}
        />
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke certificate?</DialogTitle>
            <DialogDescription>
              This will revoke certificate serial {cert.serial_number}.
            </DialogDescription>
          </DialogHeader>

          <label className="block text-sm">
            Reason
            <input
              className="mt-1 w-full rounded-md border px-3 py-2"
              value={reason}
              onChange={(event) => setReason(event.target.value)}
            />
          </label>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <DialogFooter>
            <DialogClose render={<Button variant="outline">Cancel</Button>} />
            <Button
              variant="destructive"
              onClick={onConfirm}
              disabled={loading || !reason}
            >
              {loading ? "Revoking..." : "Confirm revoke"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      {successMessage && (
        <p className="mt-2 text-sm text-green-700">{successMessage}</p>
      )}
    </>
  );
}
