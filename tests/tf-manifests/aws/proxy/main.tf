terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0, != 5.71.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  proxy_image_name = "al2023-ami-2023.5.20241001.1-kernel-6.1-x86_64"
}

data "aws_ami" "proxy_img" {
  most_recent = true

  filter {
    name   = "name"
    values = [local.proxy_image_name]
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id
}
resource "aws_security_group" "proxy_access" {
  name        = "proxy-access-sg"
  description = "Security group for proxy access"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    protocol    = "tcp"
    to_port     = 8080
    cidr_blocks = [data.aws_vpc.selected.cidr_block]
  }
}
resource "tls_private_key" "proxy_ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "generated_key" {
  key_name   = "imported-proxy-key-${var.key_pair_id != null ? var.key_pair_id : timestamp()}"
  public_key = tls_private_key.proxy_ssh_key.public_key_openssh
}


resource "aws_instance" "proxies" {
  count                       = var.proxy_count
  ami                         = data.aws_ami.proxy_img.image_id
  instance_type               = "t3.medium"
  key_name                    = aws_key_pair.generated_key.key_name
  subnet_id                   = var.subnet_public_id
  vpc_security_group_ids      = [aws_security_group.proxy_access.id]
  associate_public_ip_address = true

  tags = {
    Name = "tf-proxy-${count.index}"
  }

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = "ec2-user"
      private_key = tls_private_key.proxy_ssh_key.private_key_pem
      host        = self.public_ip
      agent       = false # Disable agent forwarding
      timeout     = "5m"
    }

    inline = [
      "sleep 30",
      "sudo yum install -y wget",
      "wget https://snapshots.mitmproxy.org/7.0.2/mitmproxy-7.0.2-linux.tar.gz",
      "mkdir mitm",
      "tar zxvf mitmproxy-7.0.2-linux.tar.gz -C mitm",
      "cd mitm",
      "nohup ./mitmdump --showhost --ssl-insecure --ignore-hosts quay.io registry.redhat.io amazonaws.com > mitm.log 2>&1 &",
      "sleep 5", # Add a delay to allow mitmdump to start
      # Generate the file locally
      "http_proxy=127.0.0.1:8080 curl -s http://mitm.it/cert/pem > ~/mitm-ca.pem"
    ]
  }

  provisioner "local-exec" {
    command = <<-EOT
            echo '${tls_private_key.proxy_ssh_key.private_key_pem}' > /tmp/private_key.pem
            chmod 600 /tmp/private_key.pem
            scp -i /tmp/private_key.pem -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ec2-user@${self.public_ip}:~/mitm-ca.pem ${var.trust_bundle_path}
        EOT
  }
}
data "local_file" "additional_trust_bundle" {
  depends_on = [aws_instance.proxies]
  filename   = var.trust_bundle_path
}

