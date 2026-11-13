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

import { Info } from "lucide-react";
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
import { Tooltip, TooltipContent, TooltipTrigger } from "@/shared/ui/tooltip";
import { useGenerateCertificate } from "../queries";
import type { GeneratedCertificate } from "../schemas";
import { CERTIFICATE_OPTIONS, LIFETIME_CHOICES, formatUnixSeconds } from "../status";

const REFUSAL_MESSAGES: Record<string, string> = {
    user_not_active: "Your account is not active. Contact your site admin.",
    no_cluster_account: "You do not have an account on this cluster yet.",
    cluster_account_not_ready: "Your cluster account is not ready yet. Try again once it is set up.",
    no_active_allocation: "You need an active allocation to get a certificate.",
    reserved_username: "Certificates cannot be issued for this username. Contact your site admin.",
    invalid_public_key: "This does not look like an SSH public key. Paste the one line from your .pub file.",
    key_type_not_allowed: "This site does not accept that key type.",
    too_many_certificates: "You have generated as many certificates as allowed for now. Use one you already have, or try again later.",
    signer_unavailable: "Certificate service is not reachable. No certificate was issued. Try again in a minute.",
    invalid_ttl: "That lifetime is longer than this site allows.",
    invalid_request: "Check the addresses. Use addresses like 192.0.2.10 or ranges like 192.0.2.0/24.",
    principal_denied: "Your cluster account cannot log in yet. Contact your site admin.",
};

const LONGEST_LIFETIME = Math.max(...LIFETIME_CHOICES.map((choice) => choice.seconds));

// The signer takes a comma-separated list with no spaces.
function addressList(text: string): string {
  return text
    .split(/[\s,]+/)
    .filter(Boolean)
    .join(",");
}

function grantSummary(result: GeneratedCertificate): string | null {
  if (!result.granted_extensions) return null;
  const allowed = CERTIFICATE_OPTIONS.filter((o) => result.granted_extensions?.includes(o.name));
  const allows =
    allowed.length > 0
      ? `Allows ${allowed.map((o) => o.label).join(", ")}.`
      : "Allows no extra options.";
  return result.force_command ? `${allows} Runs only ${result.force_command}.` : allows;
}

function OptionField({
  id,
  label,
  info,
  children,
}: {
  id: string;
  label: string;
  info: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-1">
        <span id={`${id}-label`} className="text-sm font-medium">
          {label}
        </span>
        <Tooltip>
          <TooltipTrigger
            aria-label={`About ${label.toLowerCase()}`}
            className="inline-flex h-5 w-5 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <Info className="h-3.5 w-3.5" aria-hidden="true" />
          </TooltipTrigger>
          <TooltipContent side="right">{info}</TooltipContent>
        </Tooltip>
      </div>
      <fieldset aria-labelledby={`${id}-label`} className="min-w-0 space-y-2 text-sm">
        {children}
      </fieldset>
    </div>
  );
}

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
  const [lifetime, setLifetime] = React.useState(LONGEST_LIFETIME);
  const [onlyAddresses, setOnlyAddresses] = React.useState(false);
  const [addresses, setAddresses] = React.useState("");
  const [onlyCommand, setOnlyCommand] = React.useState(false);
  const [command, setCommand] = React.useState("");
  const [blocked, setBlocked] = React.useState<string[]>([]);
  const { reset } = generate;

  React.useEffect(() => {
    if (!open) {
      setPublicKey("");
      setLifetime(LONGEST_LIFETIME);
      setOnlyAddresses(false);
      setAddresses("");
      setOnlyCommand(false);
      setCommand("");
      setBlocked([]);
      reset();
    }
  }, [open, reset]);

  // Checked here so a pasted private key is never sent anywhere.
  const isPrivateKey = publicKey.includes("PRIVATE KEY") || publicKey.includes("PuTTY-User-Key-File");
  const missingValue =
    (onlyAddresses && !addressList(addresses)) || (onlyCommand && !command.trim());
  const result = generate.data;
  const summary = result ? grantSummary(result) : null;

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!generate.isPending) onOpenChange(next);
      }}
    >
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        {result ? (
          <>
            <DialogHeader>
              <DialogTitle>Your certificate is ready</DialogTitle>
              <DialogDescription>
                Certificate {result.serial_number} for {result.principal}, valid until{" "}
                {formatUnixSeconds(result.valid_before)}. It is shown only now. If you lose it,
                generate another.
              </DialogDescription>
            </DialogHeader>
            <div className="min-w-0 space-y-4 text-sm">
              {summary ? <p>{summary}</p> : null}
              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <p className="font-medium">Certificate</p>
                  <CopyValue value={result.certificate} label="certificate" explicit>
                    Copy
                  </CopyValue>
                </div>
                <pre className="max-h-32 overflow-y-auto whitespace-pre-wrap break-all rounded-md border bg-muted p-2 font-mono text-xs">
                  {result.certificate}
                </pre>
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
                if (!publicKey.trim() || isPrivateKey || missingValue) return;
                generate.mutate({
                  public_key: publicKey.trim(),
                  ttl_seconds: lifetime,
                  exclude_extensions: blocked,
                  ...(onlyAddresses ? { source_address: addressList(addresses) } : {}),
                  ...(onlyCommand ? { force_command: command.trim() } : {}),
                });
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
                    : "Paste your .pub file, for example ~/.ssh/id_ed25519.pub. Never your private key."}
                </p>
              </div>
              {isPrivateKey ? null : (
                <div className="space-y-1 text-xs text-muted-foreground">
                  <p>No key yet? Run this on your computer:</p>
                  <div className="rounded-md border bg-muted px-3 py-2">
                    <CopyValue
                      value="ssh-keygen -t ed25519"
                      label="ssh-keygen command"
                      explicit
                      className="w-full justify-between"
                    />
                  </div>
                </div>
              )}

              <OptionField
                id="lifetime"
                label="Lifetime"
                info="How long the certificate works."
              >
                <select
                  aria-labelledby="lifetime-label"
                  value={lifetime}
                  onChange={(e) => setLifetime(Number(e.target.value))}
                  disabled={generate.isPending}
                  className="h-9 w-full rounded-md border bg-background px-3 text-sm"
                >
                  {LIFETIME_CHOICES.map((choice) => (
                    <option key={choice.seconds} value={choice.seconds}>
                      {choice.label}
                    </option>
                  ))}
                </select>
              </OptionField>

              <OptionField
                id="log-in-from"
                label="Log in from"
                info="Written into the certificate as source-address. Login nodes refuse the certificate from any other address."
              >
                <label className="flex items-center gap-2">
                  <input
                    type="radio"
                    name="log-in-from"
                    checked={!onlyAddresses}
                    onChange={() => setOnlyAddresses(false)}
                  />
                  Any address
                </label>
                <label className="flex items-center gap-2">
                  <input
                    type="radio"
                    name="log-in-from"
                    checked={onlyAddresses}
                    onChange={() => setOnlyAddresses(true)}
                  />
                  Only these addresses
                </label>
                {onlyAddresses ? (
                  <input
                    aria-label="Addresses"
                    value={addresses}
                    onChange={(e) => setAddresses(e.target.value)}
                    placeholder="192.0.2.10, 198.51.100.0/24"
                    className="h-9 w-full rounded-md border bg-background px-3 font-mono text-xs"
                  />
                ) : null}
              </OptionField>

              <OptionField
                id="command"
                label="Command"
                info="Written into the certificate as force-command. Every login runs this command and nothing else."
              >
                <label className="flex items-center gap-2">
                  <input
                    type="radio"
                    name="command"
                    checked={!onlyCommand}
                    onChange={() => setOnlyCommand(false)}
                  />
                  Normal login
                </label>
                <label className="flex items-center gap-2">
                  <input
                    type="radio"
                    name="command"
                    checked={onlyCommand}
                    onChange={() => setOnlyCommand(true)}
                  />
                  Only this command
                </label>
                {onlyCommand ? (
                  <input
                    aria-label="Command"
                    value={command}
                    onChange={(e) => setCommand(e.target.value)}
                    placeholder="/usr/bin/rsync --server ..."
                    className="h-9 w-full rounded-md border bg-background px-3 font-mono text-xs"
                  />
                ) : null}
              </OptionField>

              <OptionField
                id="allow"
                label="Allow on this certificate"
                info="Unticked options are left out of the certificate, so the login node blocks them."
              >
                <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
                  {CERTIFICATE_OPTIONS.map((option) => (
                    <label key={option.name} className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={!blocked.includes(option.name)}
                        onChange={(e) =>
                          setBlocked((current) =>
                            e.target.checked
                              ? current.filter((n) => n !== option.name)
                              : [...current, option.name],
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </OptionField>

              {generate.error ? (
                <p className="text-sm text-destructive">{refusalMessage(generate.error)}</p>
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
                <Button
                  type="submit"
                  disabled={generate.isPending || !publicKey.trim() || isPrivateKey || missingValue}
                >
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
