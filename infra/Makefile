TERRAFORM ?= terraform
TFVARS ?= terraform.tfvars

.PHONY: init plan apply

init:
	$(TERRAFORM) init

plan:
	$(TERRAFORM) plan -var-file="$(TFVARS)"

apply:
	$(TERRAFORM) apply -var-file="$(TFVARS)"
