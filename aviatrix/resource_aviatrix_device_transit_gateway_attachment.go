package aviatrix

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aviatrix/goaviatrix"
)

func resourceAviatrixDeviceTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAviatrixDeviceTransitGatewayAttachmentCreate,
		Read:   resourceAviatrixDeviceTransitGatewayAttachmentRead,
		Delete: resourceAviatrixDeviceTransitGatewayAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"device_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Device name.",
			},
			"transit_gateway_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Aviatrix Transit Gateway name.",
			},
			"connection_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Connection name.",
			},
			"transit_gateway_bgp_asn": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "BGP AS Number for transit gateway.",
			},
			"device_bgp_asn": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "BGP AS Number for the device.",
			},
			"phase1_authentication": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "SHA-256",
				Description: "Phase 1 authentication algorithm.",
			},
			"phase1_dh_groups": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     14,
				Description: "Phase 1 Diffie-Hellman groups.",
			},
			"phase1_encryption": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "AES-256-CBC",
				Description: "Phase 1 encryption algorithm.",
			},
			"phase2_authentication": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "HMAC-SHA-256",
				Description: "Phase 2 authentication algorithm.",
			},
			"phase2_dh_groups": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     14,
				Description: "Phase 2 Diffie-Hellman groups.",
			},
			"phase2_encryption": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "AES-256-CBC",
				Description: "Phase 2 encryption algorithm.",
			},
			"enable_global_accelerator": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Enable AWS Global Accelerator",
			},
			"pre_shared_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Pre-shared Key.",
			},
			"local_tunnel_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Local tunnel IP",
			},
			"remote_tunnel_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Remote tunnel IP",
			},
		},
	}
}

func marshalDeviceTransitGatewayAttachmentInput(d *schema.ResourceData) *goaviatrix.DeviceTransitGatewayAttachment {
	return &goaviatrix.DeviceTransitGatewayAttachment{
		DeviceName:              d.Get("device_name").(string),
		TransitGatewayName:      d.Get("transit_gateway_name").(string),
		ConnectionName:          d.Get("connection_name").(string),
		RoutingProtocol:         "bgp",
		TransitGatewayBgpAsn:    strconv.Itoa(d.Get("transit_gateway_bgp_asn").(int)),
		DeviceBgpAsn:            strconv.Itoa(d.Get("device_bgp_asn").(int)),
		Phase1Authentication:    d.Get("phase1_authentication").(string),
		Phase1DHGroups:          strconv.Itoa(d.Get("phase1_dh_groups").(int)),
		Phase1Encryption:        d.Get("phase1_encryption").(string),
		Phase2Authentication:    d.Get("phase2_authentication").(string),
		Phase2DHGroups:          strconv.Itoa(d.Get("phase2_dh_groups").(int)),
		Phase2Encryption:        d.Get("phase2_encryption").(string),
		EnableGlobalAccelerator: strconv.FormatBool(d.Get("enable_global_accelerator").(bool)),
		PreSharedKey:            d.Get("pre_shared_key").(string),
		LocalTunnelIP:           d.Get("local_tunnel_ip").(string),
		RemoteTunnelIP:          d.Get("remote_tunnel_ip").(string),
	}
}

func resourceAviatrixDeviceTransitGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	attachment := marshalDeviceTransitGatewayAttachmentInput(d)

	if err := client.CreateDeviceTransitGatewayAttachment(attachment); err != nil {
		return fmt.Errorf("could not create transit gateway and device attachment: %v", err)
	}

	d.SetId(attachment.ConnectionName)
	return nil
}

func resourceAviatrixDeviceTransitGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	connectionName := d.Get("connection_name").(string)
	isImport := false
	if connectionName == "" {
		isImport = true
		id := d.Id()
		d.SetId(id)
		connectionName = id
		log.Printf("[DEBUG] Looks like an import, no device_transit_gateway_attachment connection_name received. Import Id is %s", id)
	}

	attachment := &goaviatrix.DeviceTransitGatewayAttachment{
		ConnectionName: connectionName,
	}

	attachment, err := client.GetDeviceTransitGatewayAttachment(attachment)
	if err == goaviatrix.ErrNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not find device_transit_gateway_attachment %s: %v", connectionName, err)
	}

	d.Set("device_name", attachment.DeviceName)
	d.Set("transit_gateway_name", attachment.TransitGatewayName)
	d.Set("connection_name", attachment.ConnectionName)

	transitGatewayBgpAsn, err := strconv.Atoi(attachment.TransitGatewayBgpAsn)
	if err != nil {
		return fmt.Errorf("could not convert transitGatewayBgpAsn to int: %v", err)
	}
	d.Set("transit_gateway_bgp_asn", transitGatewayBgpAsn)

	deviceBgpAsn, err := strconv.Atoi(attachment.DeviceBgpAsn)
	if err != nil {
		return fmt.Errorf("could not convert deviceBgpAsn to int: %v", err)
	}
	d.Set("device_bgp_asn", deviceBgpAsn)

	d.Set("phase1_authentication", attachment.Phase1Authentication)

	phase1DhGroups, err := strconv.Atoi(attachment.Phase1DHGroups)
	if err != nil {
		return fmt.Errorf("could not convert phase1DhGroups to int: %v", err)
	}
	d.Set("phase1_dh_groups", phase1DhGroups)

	d.Set("phase1_encryption", attachment.Phase1Encryption)
	d.Set("phase2_authentication", attachment.Phase2Authentication)

	phase2DhGroups, err := strconv.Atoi(attachment.Phase2DHGroups)
	if err != nil {
		return fmt.Errorf("could not convert phase2DhGroups to int: %v", err)
	}
	d.Set("phase2_dh_groups", phase2DhGroups)

	d.Set("phase2_encryption", attachment.Phase2Encryption)

	enableGlobalAccelerator, err := strconv.ParseBool(attachment.EnableGlobalAccelerator)
	if err != nil {
		return fmt.Errorf("could not convert enableGlobalAccelerator to bool: %v", err)
	}
	d.Set("enable_global_accelerator", enableGlobalAccelerator)

	if isImport || d.Get("local_tunnel_ip") != "" {
		d.Set("local_tunnel_ip", attachment.LocalTunnelIP)
	}
	if isImport || d.Get("remote_tunnel_ip") != "" {
		d.Set("remote_tunnel_ip", attachment.RemoteTunnelIP)
	}

	d.SetId(attachment.ConnectionName)
	return nil
}

func resourceAviatrixDeviceTransitGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	cn := d.Get("connection_name").(string)

	if err := client.DeleteDeviceAttachment(cn); err != nil {
		return fmt.Errorf("could not delete transit gateway and device attachment: %v", err)
	}

	return nil
}
