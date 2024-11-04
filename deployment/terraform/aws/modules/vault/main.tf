module "label" {
  source      = "cloudposse/label/null"
  version     = "0.25.0"
  environment = var.environment
  label_order = ["namespace", "name", "environment"]
  name        = var.name
  namespace   = var.namespace
  tags        = var.tags
}

resource "aws_iam_role" "instance_role" {
  name               = "${var.namespace}-${var.name}"
  description        = "Vault role for ${module.label.id}"
  assume_role_policy = data.aws_iam_policy_document.instance_role.json
}

data "aws_iam_policy_document" "instance_role" {
  statement {
    effect  = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_instance_profile" "vault" {
  name_prefix = "${var.namespace}-${var.name}"
  role        = aws_iam_role.instance_role.name
  tags        = module.label.tags
}

resource "aws_iam_role_policy" "auto_join" {
  name   = "${var.namespace}-${var.name}-auto-join"
  policy = data.aws_iam_policy_document.auto_join.json
  role   = aws_iam_role.instance_role.id
}

data "aws_iam_policy_document" "auto_join" {
  statement {
    effect  = "Allow"
    actions = [
      "ec2:DescribeInstances",
      "ec2:DescribeTags"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "auto_unseal" {
  name   = "${var.namespace}-${var.name}-auto-unseal"
  policy = data.aws_iam_policy_document.auto_unseal.json
  role   = aws_iam_role.instance_role.id
}

resource "aws_kms_key" "vault" {
  description             = "AWS KMS key used for Vault auto-unseal and encryption"
  key_usage               = "ENCRYPT_DECRYPT"
  deletion_window_in_days = 7
  is_enabled              = true
  tags                    = merge(
    var.tags,
    {
      "Name" = "${var.namespace}-${var.name}-key"
    }
  )
}

data "aws_iam_policy_document" "auto_unseal" {
  statement {
    effect  = "Allow"
    actions = [
      "kms:DescribeKey",
      "kms:Encrypt",
      "kms:Decrypt",
    ]
    resources = [
      aws_kms_key.vault.arn,
    ]
  }
}

resource "aws_iam_role_policy" "secrets_manager" {
  name   = "${var.namespace}-${var.name}-secrets-manager"
  policy = data.aws_iam_policy_document.secrets_manager.json
  role   = aws_iam_role.instance_role.id
}

data "aws_iam_policy_document" "secrets_manager" {
  statement {
    effect  = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
    ]
    resources = [
      var.secrets_manager_arn,
    ]
  }
}

module "alb" {
  source                                  = "cloudposse/alb/aws"
  version                                 = "1.11.1"
  alb_access_logs_s3_bucket_force_destroy = var.alb_destroy_log_bucket
  attributes                              = ["alb"]
  certificate_arn                         = var.alb_certificate_arn
  deletion_protection_enabled             = var.deletion_protection
  health_check_interval                   = 60
  health_check_path                       = var.alb_health_check_path
  health_check_timeout                    = 10
  http_ingress_cidr_blocks                = var.http_ingress_cidr_blocks
  http_redirect                           = var.http_redirect
  https_enabled                           = true
  https_ingress_cidr_blocks               = var.https_ingress_cidr_blocks
  internal                                = true
  lifecycle_rule_enabled                  = true
  name                                    = module.label.id
  subnet_ids                              = var.private_subnet_ids
  tags                                    = module.label.tags
  target_group_name                       = module.label.id
  target_group_port                       = var.container_port
  target_group_target_type                = "instance"
  vpc_id                                  = var.vpc_id
  stickiness                              = var.stickiness
}

resource "aws_security_group" "vault_private" {
  name   = "${var.namespace}-${var.name}"
  description = "Vault Security Group"
  vpc_id = var.vpc_id
  tags   = merge(
    var.tags,
    {
      "Name" = "${var.namespace}-${var.name}-sg"
    }
  )
}

resource "aws_security_group_rule" "vault_internal_api" {
  security_group_id = aws_security_group.vault_private.id
  description       = "Vault Internal API"
  type              = "ingress"
  from_port         = 8200
  to_port           = 8200
  protocol          = "tcp"
  self              = true
}

resource "aws_security_group_rule" "vault_internal_raft" {
  security_group_id = aws_security_group.vault_private.id
  description       = "Vault Raft"
  type              = "ingress"
  from_port         = 8201
  to_port           = 8201
  protocol          = "tcp"
  self              = true
}

resource "aws_security_group_rule" "vault_ssh_inbound" {
  description       = "Allow SSH"
  security_group_id = aws_security_group.vault_private.id
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  self              = true
}

resource "aws_security_group_rule" "vault_alb_inbound" {
  security_group_id        = aws_security_group.vault_private.id
  description              = "Allow AWS ALB to reach Vault nodes"
  type                     = "ingress"
  from_port                = 8200
  to_port                  = 8200
  protocol                 = "tcp"
  source_security_group_id = module.alb.security_group_id
}

resource "aws_security_group_rule" "vault_outbound" {
  security_group_id = aws_security_group.vault_private.id
  description       = "Allow Vault nodes to send outbound traffic"
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_launch_template" "vault" {
  name          = "${var.namespace}-${var.name}"
  image_id      = var.ubuntu_ami
  instance_type = var.instance_type
  key_name      = var.ssh_key_name
  user_data     = base64encode(templatefile("${path.module}/templates/install_vault_script.sh.tpl",
    {
      region                = var.region
      name                  = var.namespace
      vault_version         = var.vault_version
      kms_key_arn           = aws_kms_key.vault.arn
      secrets_manager_arn   = var.secrets_manager_arn
      leader_tls_servername = var.leader_tls_servername
    }))
  vpc_security_group_ids = [
    aws_security_group.vault_private.id,
  ]

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_type           = "gp3"
      volume_size           = 100
      throughput            = 150
      iops                  = 3000
      delete_on_termination = true
    }
  }

  iam_instance_profile {
    name = aws_iam_instance_profile.vault.name
  }

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }
}

resource "aws_autoscaling_group" "vault" {
  name                = "${var.namespace}-${var.name}-asg"
  min_size            = var.min_nodes
  max_size            = var.max_nodes
  desired_capacity    = var.min_nodes
  vpc_zone_identifier = var.private_subnet_ids
  target_group_arns   = [module.alb.default_target_group_arn]

  launch_template {
    id      = aws_launch_template.vault.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "${var.namespace}-${var.name}-server"
    propagate_at_launch = true
  }
}