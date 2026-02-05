output "vpc_id" {
  description = "Created VPC ID"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "Created public subnet IDs"
  value       = module.vpc.public_subnets
}
