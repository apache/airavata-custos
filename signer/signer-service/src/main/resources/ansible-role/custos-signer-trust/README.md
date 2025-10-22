# Custos Signer Trust Ansible Role

This Ansible role configures SSH servers to trust Custos SSH Certificate Authority (CA) and automatically fetch Key Revocation Lists (KRLs).

## Purpose

The role automates the configuration of SSH servers to:
- Trust Custos CA public keys for certificate-based authentication
- Automatically fetch and apply SSH Key Revocation Lists (KRLs)
- Reload SSH daemon configuration without service interruption

## Requirements

- OpenSSH 7.2 or later
- Root or sudo access on target systems
- Network access to Custos Signer Service
- curl command available

## Role Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `tenant_id` | Custos tenant identifier | `airavata-prod` |
| `client_id` | Custos client identifier | `airavata-hpcA` |
| `ca_public_key` | CA public key content | `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5...` |
| `signer_service_url` | Base URL of Custos Signer Service | `https://signer.custos.org` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `krl_fetch_interval_minutes` | KRL fetch interval in minutes | `5` |
| `ssh_config_backup` | Create backup of sshd_config | `true` |
| `validate_ssh_config` | Validate sshd configuration | `true` |

## Example Playbook

```yaml
---
- hosts: compute_nodes
  become: yes
  roles:
    - role: custos-signer-trust
      vars:
        tenant_id: airavata-prod
        client_id: airavata-hpcA
        ca_public_key: "{{ lookup('file', 'ca-keys/hpcA.pub') }}"
        signer_service_url: https://signer.custos.org
```

## What This Role Does

1. **Creates CA Keys Directory**: `/etc/ssh/ca_keys/`
2. **Installs CA Public Key**: Places CA public key in `/etc/ssh/ca_keys/{tenant_id}_{client_id}.pub`
3. **Configures SSH Daemon**: Updates `/etc/ssh/sshd_config` with:
   - `TrustedUserCAKeys /etc/ssh/ca_keys/*.pub`
   - `RevokedKeys /etc/ssh/revoked.krl`
4. **Deploys KRL Fetch Script**: Creates `/usr/local/bin/custos-fetch-krl.sh`
5. **Schedules KRL Updates**: Sets up cron job to fetch KRL every 5 minutes
6. **Validates Configuration**: Runs `sshd -t` to validate configuration
7. **Reloads SSH Daemon**: Uses `systemctl reload sshd` (not restart)

## Files Created

- `/etc/ssh/ca_keys/{tenant_id}_{client_id}.pub` - CA public key
- `/etc/ssh/revoked.krl` - SSH Key Revocation List
- `/usr/local/bin/custos-fetch-krl.sh` - KRL fetch script
- `/var/log/custos-krl-fetch.log` - KRL fetch logs

## Testing

After applying the role, test certificate authentication:

```bash
# Test CA public key installation
ssh-keygen -Lf /etc/ssh/ca_keys/*.pub

# Test certificate authentication (if you have a valid certificate)
ssh -i /path/to/certificate user@hostname
```

## Troubleshooting

### Check KRL Fetch Logs
```bash
tail -f /var/log/custos-krl-fetch.log
```

### Manual KRL Fetch
```bash
/usr/local/bin/custos-fetch-krl.sh
```

### Validate SSH Configuration
```bash
sshd -t
```

### Check SSH Daemon Status
```bash
systemctl status sshd
```

## Security Notes

- The role uses `systemctl reload sshd` instead of restart to avoid service interruption
- CA public keys are installed with restrictive permissions (644)
- KRL fetch script validates downloaded content before applying
- SSH configuration is backed up before modification
- All operations are logged for audit purposes

## Revocation

To revoke certificates:
1. Use Custos Signer Service revocation API
2. KRL will be automatically updated within 5 minutes
3. SSH daemon will be reloaded to apply new KRL
4. Revoked certificates will be rejected immediately
