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
			// TODO Add validation that service_principal is set iff mount_type == ServicePrincipal
			// TODO - is service_principal the right name??
			"service_principal": {
				Type:     schema.TypeList,
				Required: true,
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

func resourceAzureAdlsGen2Create(d *schema.ResourceData, m interface{}) error {
	client := m.(service.DBApiClient)
	clusterID := d.Get("cluster_id").(string)
	err := changeClusterIntoRunningState(clusterID, client)
	if err != nil {
		return err
	}
	containerName := d.Get("container_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	directory := d.Get("directory").(string)
	mountName := d.Get("mount_name").(string)
	mountType := d.Get("mount_type").(string)
	initializeFileSystem := d.Get("initialize_file_system").(bool)

	var adlsGen2Mount *AzureADLSGen2Mount
	var servicePrincipal map[string]interface{}
	switch mountType {
	case "AADPassthrough":
		adlsGen2Mount = NewAzureADLSGen2MountAADPassthrough(containerName, storageAccountName, directory, mountName, initializeFileSystem)
	case "ServicePrincipal":
		servicePrincipalList := d.Get("service_principal").([]interface{})
		if len(servicePrincipalList) == 0 {
			return fmt.Errorf("Error: when mount_type is ServicePrincipal, service_principal block is required")
		}
		servicePrincipal := servicePrincipalList[0].(map[string]interface{})
		tenantID := servicePrincipal["tenant_id"].(string)
		clientID := servicePrincipal["client_id"].(string)
		clientSecretScope := servicePrincipal["client_secret_scope"].(string)
		clientSecretKey := servicePrincipal["client_secret_key"].(string)
		adlsGen2Mount = NewAzureADLSGen2MountServicePrincipal(containerName, storageAccountName, directory, mountName, clientID, tenantID, clientSecretScope, clientSecretKey, initializeFileSystem)

		servicePrincipal = map[string]interface{}{
			"tenant_id":           tenantID,
			"client_id":           clientID,
			"client_secret_scope": clientSecretScope,
			"client_secret_key":   clientSecretKey,
		}
	default:
		return fmt.Errorf("Unsupported value for mount_type: '%s'", mountType)
	}

	err = adlsGen2Mount.Create(client, clusterID)
	if err != nil {
		return err
	}
	d.SetId(mountName)

	err = d.Set("cluster_id", clusterID)
	if err != nil {
		return err
	}
	err = d.Set("mount_name", mountName)
	if err != nil {
		return err
	}
	err = d.Set("mount_type", mountType)
	if err != nil {
		return err
	}
	if servicePrincipal != nil {
		err = d.Set("service_principal", servicePrincipal)
		if err != nil {
			return err
		}
	}
	err = d.Set("initialize_file_system", initializeFileSystem)
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
	containerName := d.Get("container_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	directory := d.Get("directory").(string)
	mountName := d.Get("mount_name").(string)
	useAADPassthrough := d.Get("use_aad_passthrough").(bool)
	tenantID := d.Get("tenant_id").(string)
	clientID := d.Get("client_id").(string)
	clientSecretScope := d.Get("client_secret_scope").(string)
	clientSecretKey := d.Get("client_secret_key").(string)
	initializeFileSystem := d.Get("initialize_file_system").(bool)

	adlsGen2Mount := NewAzureADLSGen2Mount(containerName, storageAccountName, directory, mountName, useAADPassthrough, clientID, tenantID,
		clientSecretScope, clientSecretKey, initializeFileSystem)

	url, err := adlsGen2Mount.Read(client, clusterID)
	if err != nil {
		//Reset id in case of inability to find mount
		if strings.Contains(err.Error(), "Unable to find mount point!") ||
			strings.Contains(err.Error(), fmt.Sprintf("/mnt/%s does not exist.", mountName)) {
			d.SetId("")
			return nil
		}
		return err
	}
	container, storageAcc, dir, err := ProcessAzureWasbAbfssUris(url)
	if err != nil {
		return err
	}
	err = d.Set("container_name", container)
	if err != nil {
		return err
	}
	err = d.Set("storage_account_name", storageAcc)
	if err != nil {
		return err
	}
	err = d.Set("directory", dir)
	return err
}

func resourceAzureAdlsGen2Delete(d *schema.ResourceData, m interface{}) error {
	client := m.(service.DBApiClient)
	clusterID := d.Get("cluster_id").(string)
	err := changeClusterIntoRunningState(clusterID, client)
	if err != nil {
		return err
	}
	containerName := d.Get("container_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	directory := d.Get("directory").(string)
	mountName := d.Get("mount_name").(string)
	useAADPassthrough := d.Get("use_aad_passthrough").(bool)
	tenantID := d.Get("tenant_id").(string)
	clientID := d.Get("client_id").(string)
	clientSecretScope := d.Get("client_secret_scope").(string)
	clientSecretKey := d.Get("client_secret_key").(string)
	initializeFileSystem := d.Get("initialize_file_system").(bool)

	adlsGen2Mount := NewAzureADLSGen2Mount(containerName, storageAccountName, directory, mountName, useAADPassthrough, clientID, tenantID,
		clientSecretScope, clientSecretKey, initializeFileSystem)
	return adlsGen2Mount.Delete(client, clusterID)
}
