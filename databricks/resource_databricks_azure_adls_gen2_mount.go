package databricks

import (
	"fmt"
	"github.com/databrickslabs/databricks-terraform/client/service"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"strings"
)

func resourceAzureAdlsGen2Mount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAzureAdlsGen2Create,
		Read:   resourceAzureAdlsGen2Read,
		Delete: resourceAzureAdlsGen2Delete,

		Schema: map[string]*schema.Schema{
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_account_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"mount_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mount_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"AADPassthrough", "ServicePrincipal"}, false),
			},
			"service_principal": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tenant_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"client_secret_scope": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"client_secret_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							ForceNew:  true,
						},
					},
				},
			},
			"initialize_file_system": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
		},
	}
}
func resourceAzureAdlsGen2GetMountFromResourceData(d *schema.ResourceData) (*AzureADLSGen2Mount, error) {
	containerName := d.Get("container_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	directory := d.Get("directory").(string)
	mountName := d.Get("mount_name").(string)
	mountType := d.Get("mount_type").(string)
	initializeFileSystem := d.Get("initialize_file_system").(bool)

	switch mountType {
	case "AADPassthrough":
		return NewAzureADLSGen2Mount(containerName, storageAccountName, directory, mountName, AzureADLSGen2MountType_AADPassthrough, nil, initializeFileSystem), nil
	case "ServicePrincipal":
		servicePrincipalList := d.Get("service_principal").([]interface{})
		if len(servicePrincipalList) == 0 {
			return nil, fmt.Errorf("Error: when mount_type is ServicePrincipal, service_principal block is required")
		}
		servicePrincipalProps := servicePrincipalList[0].(map[string]interface{})
		tenantID := servicePrincipalProps["tenant_id"].(string)
		clientID := servicePrincipalProps["client_id"].(string)
		clientSecretScope := servicePrincipalProps["client_secret_scope"].(string)
		clientSecretKey := servicePrincipalProps["client_secret_key"].(string)
		servicePrincipal := AzureADLSGen2MountServicePrincipal{
			TenantID:    tenantID,
			ClientID:    clientID,
			SecretScope: clientSecretScope,
			SecretKey:   clientSecretKey,
		}
		return NewAzureADLSGen2Mount(containerName, storageAccountName, directory, mountName, AzureADLSGen2MountType_ServicePrincipal, &servicePrincipal, initializeFileSystem), nil
	}
	return nil, fmt.Errorf("Unsupported value for mount_type: '%s'", mountType)
}

func resourceAzureAdlsGen2Create(d *schema.ResourceData, m interface{}) error {
	client := m.(service.DBApiClient)
	clusterID := d.Get("cluster_id").(string)
	err := changeClusterIntoRunningState(clusterID, client)
	if err != nil {
		return err
	}

	adlsGen2Mount, err := resourceAzureAdlsGen2GetMountFromResourceData(d)
	if err != nil {
		return err
	}

	err = adlsGen2Mount.Create(client, clusterID)
	if err != nil {
		return err
	}
	d.SetId(adlsGen2Mount.MountName)

	err = d.Set("cluster_id", clusterID)
	if err != nil {
		return err
	}
	err = d.Set("mount_name", adlsGen2Mount.MountName)
	if err != nil {
		return err
	}
	err = d.Set("mount_type", adlsGen2Mount.MountType)
	if err != nil {
		return err
	}
	if adlsGen2Mount.MountType == AzureADLSGen2MountType_ServicePrincipal {
		servicePrincipalMap := map[string]interface{}{
			"tenant_id":           adlsGen2Mount.ServicePrincipal.TenantID,
			"client_id":           adlsGen2Mount.ServicePrincipal.ClientID,
			"client_secret_scope": adlsGen2Mount.ServicePrincipal.SecretScope,
			"client_secret_key":   adlsGen2Mount.ServicePrincipal.SecretKey,
		}
		err = d.Set("service_principal", []interface{}{servicePrincipalMap})
		if err != nil {
			return err
		}
	}
	err = d.Set("initialize_file_system", adlsGen2Mount.InitializeFileSystem)
	if err != nil {
		return err
	}

	return resourceAzureAdlsGen2Read(d, m)
}
func resourceAzureAdlsGen2Read(d *schema.ResourceData, m interface{}) error {
	client := m.(service.DBApiClient)
	clusterID := d.Get("cluster_id").(string)
	err := changeClusterIntoRunningState(clusterID, client)
	if err != nil {
		return err
	}

	adlsGen2Mount, err := resourceAzureAdlsGen2GetMountFromResourceData(d)
	if err != nil {
		return err
	}

	url, err := adlsGen2Mount.Read(client, clusterID)
	if err != nil {
		//Reset id in case of inability to find mount
		if strings.Contains(err.Error(), "Unable to find mount point!") ||
			strings.Contains(err.Error(), fmt.Sprintf("/mnt/%s does not exist.", adlsGen2Mount.MountName)) {
			d.SetId("")
			return nil
		}
		return err
	}
	containerName, storageAccount, directory, err := ProcessAzureWasbAbfssUris(url)
	if err != nil {
		return err
	}
	err = d.Set("container_name", containerName)
	if err != nil {
		return err
	}
	err = d.Set("storage_account_name", storageAccount)
	if err != nil {
		return err
	}
	err = d.Set("directory", directory)
	return err
}

func resourceAzureAdlsGen2Delete(d *schema.ResourceData, m interface{}) error {
	client := m.(service.DBApiClient)
	clusterID := d.Get("cluster_id").(string)
	err := changeClusterIntoRunningState(clusterID, client)
	if err != nil {
		return err
	}

	adlsGen2Mount, err := resourceAzureAdlsGen2GetMountFromResourceData(d)
	if err != nil {
		return err
	}

	return adlsGen2Mount.Delete(client, clusterID)
}
