variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.20.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones to use for the public subnets (must align with public_subnet_cidrs length)"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]

  validation {
    condition     = length(var.availability_zones) == length(var.public_subnet_cidrs)
    error_message = "availability_zones length must match public_subnet_cidrs length."
  }
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets (one per AZ)"
  type        = list(string)
  default     = ["10.20.0.0/24", "10.20.1.0/24"]

  validation {
    condition     = length(var.public_subnet_cidrs) > 0
    error_message = "public_subnet_cidrs must include at least one subnet."
  }
}

variable "map_public_ip_on_launch" {
  description = "Whether instances launched in public subnets receive a public IP by default (set true to simulate CCC.VPC.CN02 failure; set false to satisfy CCC.VPC.CN02)."
  type        = bool
  default     = true
}
