module github.com/databrickslabs/databricks-terraform

go 1.13

require (
	github.com/google/go-querystring v1.0.0
	github.com/hashicorp/go-getter v1.4.2-0.20200106182914-9813cbd4eb02 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0 // indirect
	github.com/hashicorp/terraform-config-inspect v0.0.0-20191212124732-c6ae6269b9d7 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.11.0
	github.com/joho/godotenv v1.3.0
	github.com/r3labs/diff v0.0.0-20191120142937-b4ed99a31f5a
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/terraform-providers/terraform-provider-azuread v0.8.0
	github.com/terraform-providers/terraform-provider-azurerm v1.44.1-0.20200409013256-fc0b9df8ef98
)

replace github.com/Azure/go-autorest => github.com/tombuildsstuff/go-autorest v14.0.1-0.20200317095413-f2d2d0252c3c+incompatible

replace github.com/Azure/go-autorest/autorest => github.com/tombuildsstuff/go-autorest/autorest v0.10.1-0.20200317095413-f2d2d0252c3c

replace github.com/Azure/go-autorest/autorest/azure/auth => github.com/tombuildsstuff/go-autorest/autorest/azure/auth v0.4.3-0.20200317095413-f2d2d0252c3c
