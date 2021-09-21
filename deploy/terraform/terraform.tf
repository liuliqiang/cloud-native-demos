terraform {
  required_providers {
    vultr = {
      source = "vultr/vultr"
      version = "2.4.2"
    }
  }
}

variable "vultr_api_key" {
  description = "key for vultr api access"
  type        = string
  default     = ""
}

# Configure the Vultr Provider
provider "vultr" {
  api_key = var.vultr_api_key
  rate_limit = 600
  retry_limit = 1
}

resource "vultr_instance" "my_instance" {
    plan = "vc2-1c-1gb"
    region = "sgp"
    os_id = "387"
    label = "terraform"
    tag = "terraform,hashicorp"
    hostname = "terraform-test"
    backups = "disabled"
    activation_email = false
}