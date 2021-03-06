package aviatrix

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aviatrix/goaviatrix"
)

func resourceAviatrixFirewallInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAviatrixFirewallInstanceCreate,
		Read:   resourceAviatrixFirewallInstanceRead,
		Delete: resourceAviatrixFirewallInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Security VPC.",
			},
			"firenet_gw_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the primary FireNet gateway.",
			},
			"firewall_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the firewall instance to be created.",
			},
			"firewall_image": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "One of the AWS AMIs from Palo Alto Networks.",
			},
			"firewall_size": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Instance size of the firewall.",
			},
			"egress_subnet": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Egress Interface Subnet.",
			},
			"management_subnet": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Management Interface Subnet.",
			},
			"firewall_image_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Version of firewall image.",
			},
			"key_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The .pem file name for SSH access to the firewall instance.",
			},
			"iam_role": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "IAM role.",
			},
			"bootstrap_bucket_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Bootstrap bucket name.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Applicable to Azure deployment only. 'admin' as a username is not accepted.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Applicable to Azure deployment only.",
			},
			"zone": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAzureAZ,
				Description:  "Availability Zone. Only available for AZURE. Must be in the form 'az-n', for example, 'az-2'.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the firewall instance created.",
			},
			"lan_interface": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of Lan Interface created.",
			},
			"management_interface": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of Management Interface created.",
			},
			"egress_interface": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of Egress Interface created.",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Management Public IP.",
			},
		},
	}
}

func resourceAviatrixFirewallInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	firewallInstance := &goaviatrix.FirewallInstance{
		VpcID:                d.Get("vpc_id").(string),
		GwName:               d.Get("firenet_gw_name").(string),
		FirewallName:         d.Get("firewall_name").(string),
		FirewallImage:        d.Get("firewall_image").(string),
		FirewallImageVersion: d.Get("firewall_image_version").(string),
		FirewallSize:         d.Get("firewall_size").(string),
		EgressSubnet:         d.Get("egress_subnet").(string),
		ManagementSubnet:     d.Get("management_subnet").(string),
		KeyName:              d.Get("key_name").(string),
		IamRole:              d.Get("iam_role").(string),
		BootstrapBucketName:  d.Get("bootstrap_bucket_name").(string),
		Username:             d.Get("username").(string),
		Password:             d.Get("password").(string),
	}

	cloudType, err := client.GetVpcCloudTypeById(firewallInstance.VpcID)
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			return fmt.Errorf("could not find the vpc with vpc_id=%s: %v", firewallInstance.VpcID, err)
		}
		return fmt.Errorf("could get the cloud type from the vpc_id=%s: %v", firewallInstance.VpcID, err)
	}
	zone := d.Get("zone").(string)
	if zone != "" && cloudType != goaviatrix.AZURE {
		return fmt.Errorf("'zone' attribute is only valid for AZURE")
	}
	if zone != "" {
		firewallInstance.EgressSubnet = fmt.Sprintf("%s~~%s~~", firewallInstance.EgressSubnet, zone)
		firewallInstance.ManagementSubnet = fmt.Sprintf("%s~~%s~~", firewallInstance.ManagementSubnet, zone)
	}

	instanceID, err := client.CreateFirewallInstance(firewallInstance)
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			return fmt.Errorf("failed to get firewall instance information")
		}
		return fmt.Errorf("failed to create a new firewall instance: %s", err)
	}

	d.SetId(instanceID)
	return resourceAviatrixFirewallInstanceRead(d, meta)
}

func resourceAviatrixFirewallInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	instanceID := d.Get("instance_id").(string)
	var isImport bool
	if instanceID == "" {
		id := d.Id()
		log.Printf("[DEBUG] Looks like an import, no firewall names received. Import Id is %s", id)
		d.Set("instance_id", id)
		d.SetId(id)
		isImport = true
	}

	firewallInstance := &goaviatrix.FirewallInstance{
		InstanceID: d.Get("instance_id").(string),
	}

	fI, err := client.GetFirewallInstance(firewallInstance)
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("couldn't find Firewall Instance: %s", err)
	}

	log.Printf("[INFO] Found Firewall Instance: %#v", firewallInstance)

	d.Set("vpc_id", fI.VpcID)
	d.Set("firenet_gw_name", fI.GwName)
	d.Set("firewall_name", fI.FirewallName)
	d.Set("firewall_image", fI.FirewallImage)
	d.Set("firewall_size", fI.FirewallSize)
	d.Set("instance_id", fI.InstanceID)
	d.Set("egress_subnet", fI.EgressSubnet)
	d.Set("management_subnet", fI.ManagementSubnet)

	if (d.Get("zone").(string) != "" || isImport) && fI.AvailabilityZone != "AvailabilitySet" &&
		fI.AvailabilityZone != "" && fI.CloudVendor == "Azure ARM" {
		d.Set("zone", "az-"+fI.AvailabilityZone)
	}

	d.Set("lan_interface", fI.LanInterface)
	d.Set("management_interface", fI.ManagementInterface)
	d.Set("egress_interface", fI.EgressInterface)
	d.Set("public_ip", fI.ManagementPublicIP)

	if fI.FirewallImageVersion != "" {
		d.Set("firewall_image_version", fI.FirewallImageVersion)
	}
	if d.Get("key_name") != "" {
		d.Set("key_name", fI.KeyName)
	}
	if fI.IamRole != "" {
		d.Set("iam_role", fI.IamRole)
	}
	if fI.BootstrapBucketName != "" {
		d.Set("bootstrap_bucket_name", fI.BootstrapBucketName)
	}
	if fI.Username != "" {
		d.Set("username", fI.Username)
	}

	return nil
}

func resourceAviatrixFirewallInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	firewallInstance := &goaviatrix.FirewallInstance{
		VpcID:      d.Get("vpc_id").(string),
		InstanceID: d.Get("instance_id").(string),
	}

	log.Printf("[INFO] Deleting firewall instance: %#v", firewallInstance)

	err := client.DeleteFirewallInstance(firewallInstance)
	if err != nil {
		return fmt.Errorf("failed to delete firewall instance: %s", err)
	}

	return nil
}
