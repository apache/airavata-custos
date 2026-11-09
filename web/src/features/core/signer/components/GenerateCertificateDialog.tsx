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

import * as React from "react";
import { ApiError } from "@/shared/api/client";
import { Button } from "@/shared/ui/button";
import { CopyValue } from "@/shared/ui/CopyValue";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Label } from "@/shared/ui/label";
import { useGenerateCertificate } from "../queries";
import type { GeneratedCertificate } from "../schemas";
import { formatUnixSeconds } from "../status";

const REFUSAL_MESSAGES: Record<string, string> = {
  user_not_active: "Your account is not active. Contact your site admin.",
  no_cluster_account: "You do not have an account on this cluster yet.",
  cluster_account_not_ready: "Your cluster account is not ready yet. Try again once it is set up.",
  no_active_allocation: "You need an active allocation to get a certificate.",
  reserved_username: "Certificates cannot be issued for this username. Contact your site admin.",
  invalid_public_key: "This does not look like an SSH public key. Paste the one line from your .pub file.",
  key_type_not_allowed: "This site does not accept that key type.",
  signer_unavailable: "Certificate service is not reachable. No certificate was issued. Try again in a minute.",
};

function refusalMessage(error: unknown): string {
  const code = error instanceof ApiError ? (error.body as { error?: string } | null)?.error : undefined;
  return (code && REFUSAL_MESSAGES[code]) || "Could not generate a certificate. Please try again.";
}

function keyFile(certificate: string): string {
  if (certificate.startsWith("ssh-rsa")) return "~/.ssh/id_rsa";
  if (certificate.startsWith("ecdsa")) return "~/.ssh/id_ecdsa";
  return "~/.ssh/id_ed25519";
}

function sshCommand(result: GeneratedCertificate): string {
  const port = result.login_port === 22 ? "" : `-p ${result.login_port} `;
  return `ssh -i ${keyFile(result.certificate)} ${port}${result.principal}@${result.login_host}`;
}

export type GenerateCertificateDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function GenerateCertificateDialog({ open, onOpenChange }: GenerateCertificateDialogProps) {
  const generate = useGenerateCertificate();
  const [publicKey, setPublicKey] = React.useState("");
  const { reset } = generate;

  React.useEffect(() => {
    if (!open) {
      setPublicKey("");
      reset();
    }
  }, [open, reset]);

  // Checked here so a pasted private key is never sent anywhere.
  const isPrivateKey = publicKey.includes("PRIVATE KEY") || publicKey.includes("PuTTY-User-Key-File");
  const result = generate.data;

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!generate.isPending) onOpenChange(next);
      }}
    >
      <DialogContent>
        {result ? (
          <>
            <DialogHeader>
              <DialogTitle>Your certificate is ready</DialogTitle>
              <DialogDescription>
                {result.principal}, valid until {formatUnixSeconds(result.valid_before)}. It is shown
                only now. If you lose it, generate another.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 text-sm">
              <div className="space-y-1">
                <p className="font-medium">Certificate</p>
                <CopyValue value={result.certificate} label="certificate" explicit />
              </div>
              <ol className="list-decimal space-y-2 pl-5">
                <li>
                  Save it next to your private key as{" "}
                  <code className="font-mono text-xs">{keyFile(result.certificate)}-cert.pub</code>
                </li>
                <li>
                  Run <CopyValue value={sshCommand(result)} label="ssh command" explicit />
                </li>
              </ol>
            </div>
            <DialogFooter>
              <Button onClick={() => onOpenChange(false)}>Done</Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Generate certificate</DialogTitle>
              <DialogDescription>Logs in with your own cluster account.</DialogDescription>
            </DialogHeader>
            <form
              className="space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                if (publicKey.trim() && !isPrivateKey) generate.mutate(publicKey.trim());
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="public-key">Public key</Label>
                <textarea
                  id="public-key"
                  value={publicKey}
                  onChange={(e) => setPublicKey(e.target.value)}
                  rows={4}
                  aria-invalid={isPrivateKey}
                  aria-describedby="public-key-help"
                  disabled={generate.isPending}
                  className="w-full rounded-md border border-border bg-background p-2 font-mono text-xs"
                />
                <p id="public-key-help" className="text-xs text-muted-foreground">
                  {isPrivateKey
                    ? "This is a private key. Keep it on your computer and paste the .pub file instead."
                    : "Paste your .pub file, for example ~/.ssh/id_ed25519.pub. Never your private key. No key yet? Run ssh-keygen -t ed25519 on your computer."}
                </p>
              </div>
              {generate.error ? (
                <p className="text-sm text-[color:var(--custos-red-700)]">
                  {refusalMessage(generate.error)}
                </p>
              ) : null}
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  disabled={generate.isPending}
                  onClick={() => onOpenChange(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={generate.isPending || !publicKey.trim() || isPrivateKey}>
                  {generate.isPending ? "Generating…" : "Generate certificate"}
                </Button>
              </DialogFooter>
            </form>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
