package azure

import (
	"context"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/chaosblade-io/chaosblade-exec-cloud/exec/category"
	"github.com/chaosblade-io/chaosblade-spec-go/channel"
	"github.com/chaosblade-io/chaosblade-spec-go/log"
	"github.com/chaosblade-io/chaosblade-spec-go/spec"

	"github.com/chaosblade-io/chaosblade-spec-go/util"
)

const VmBin = "chaos_azure_vm"

type VmActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewVmActionSpec() spec.ExpActionCommandSpec {
	return &VmActionSpec{
		spec.BaseExpActionCommandSpec{
			ActionFlags: []spec.ExpFlagSpec{
				&spec.ExpFlag{
					Name: "tenantId",
					Desc: "the tenantId of azure, if not provided, get from env AZURE_TENANT_ID",
				},
				&spec.ExpFlag{
					Name: "clientId",
					Desc: "the clientId of azure, if not provided, get from env AZURE_CLIENT_ID",
				},
				&spec.ExpFlag{
					Name: "clientSecret",
					Desc: "the clientSecret of azure, if not provided, get from env AZURE_CLIENT_SECRET",
				},
				&spec.ExpFlag{
					Name: "subscriptionId",
					Desc: "the subscriptionId of azure, if not provided, get from env AZURE_SUBSCRIPTION_ID",
				},
				&spec.ExpFlag{
					Name: "regionId",
					Desc: "the regionId of azure, like: AZURE, CHINA, US_GOVERMENT, GERMANY. If not provided, get from env AZURE_CLOUD ",
				},
				&spec.ExpFlag{
					Name: "resourceGroup",
					Desc: "the resourceGroup of azure, if not provided, get from env AZURE_RESOURCE_GROUP",
				},
				&spec.ExpFlag{
					Name: "type",
					Desc: "the operation of instances, support start, stop, etc",
				},
				&spec.ExpFlag{
					Name: "vmnames",
					Desc: "the virtual machines name list, split by comma",
				},
			},
			ActionExecutor: &VmExecutor{},
			ActionExample: `
# stop instances which instance id is i-x,i-y
blade create azure vm --tenantId xxx --clientId xxx --clientSecret xxx --subscriptionId xxx --resourceGroup xxx --type stop --vmnames i-x,i-y

# start instances which instance id is i-x,i-y
blade create aliyun ecs --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type start --instances i-x,i-y

# reboot instances which instance id is i-x,i-y
blade create aliyun ecs --accessKeyId xxx --accessKeySecret yyy --regionId cn-qingdao --type reboot --instances i-x,i-y`,
			ActionPrograms:   []string{VmBin},
			ActionCategories: []string{category.Cloud + "_" + category.Azure + "_" + category.VirtualMachine},
		},
	}
}

func (*VmActionSpec) Name() string {
	return "vm"
}

func (*VmActionSpec) Aliases() []string {
	return []string{}
}
func (*VmActionSpec) ShortDesc() string {
	return "do some azure virtual machines Operations, like stop, start"
}

func (b *VmActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do someazure virtual machines Operations, like stop, start"
}

type VmExecutor struct {
	channel spec.Channel
}

func (*VmExecutor) Name() string {
	return "vm"
}

var localChannel = channel.NewLocalChannel()

func (be *VmExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	if be.channel == nil {
		util.Errorf(uid, util.GetRunFuncName(), spec.ChannelNil.Msg)
		return spec.ResponseFailWithFlags(spec.ChannelNil)
	}
	tenantId := model.ActionFlags["tenantId"]
	if tenantId == "" {
		val, ok := os.LookupEnv("AZURE_TENANT_ID")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_TENANT_ID from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "tenantId")
		}
		tenantId = val
	}
	clientId := model.ActionFlags["clientId"]
	if clientId == "" {
		val, ok := os.LookupEnv("AZURE_CLIENT_ID")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_CLIENT_ID from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "clientId")
		}
		clientId = val
	}
	clientSecret := model.ActionFlags["clientSecret"]
	if clientSecret == "" {
		val, ok := os.LookupEnv("AZURE_CLIENT_SECRET")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_CLIENT_SECRET from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "clientSecret")
		}
		clientSecret = val
	}
	subscriptionId := model.ActionFlags["subscriptionId"]
	if subscriptionId == "" {
		val, ok := os.LookupEnv("AZURE_SUBSCRIPTION_ID")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_SUBSCRIPTION_ID from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "subscriptionId")
		}
		subscriptionId = val
	}
	resourceGroup := model.ActionFlags["resourceGroup"]
	if resourceGroup == "" {
		val, ok := os.LookupEnv("AZURE_RESOURCE_GROUP")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_RESOURCE_GROUP from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "resourceGroup")
		}
		resourceGroup = val
	}
	// regionId := model.ActionFlags["regionId"]
	vmnames := model.ActionFlags["vmnames"]
	if vmnames == "" {
		log.Errorf(ctx, "vmnames is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "vmnames")
	}
	operationType := model.ActionFlags["type"]
	if operationType == "" {
		log.Errorf(ctx, "operationType is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "operationType")
	}

	vmNamesArray := strings.Split(vmnames, ",")
	return be.start(ctx, tenantId, clientId, clientSecret, subscriptionId, operationType, resourceGroup, vmNamesArray)
}

func (be *VmExecutor) start(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, operationType, resourceGroup string, vmNamesArray []string) *spec.Response {
	switch operationType {
	case "start":
		return be.startInstances(ctx, tenantId, clientId, clientSecret, subscriptionId, resourceGroup, vmNamesArray)
	case "stop":
		return be.stopInstances(ctx, tenantId, clientId, clientSecret, subscriptionId, resourceGroup, vmNamesArray)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support start, stop, reboot)")
	}
	select {}
}

func (be *VmExecutor) startInstances(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, resourceGroup string, vmNamesArray []string) *spec.Response {
	client, err := CreateVmClient(tenantId, clientId, clientSecret, subscriptionId)
	if err != nil {
		log.Errorf(ctx, "create azure client failed, err: %s", err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create azure client failed")
	}
	pollers := make([]*runtime.Poller[armcompute.VirtualMachinesClientStartResponse], 0)
	for _, vmName := range vmNamesArray {
		poller, err := client.BeginStart(ctx, resourceGroup, vmName, nil)
		if err != nil {
			log.Errorf(ctx, "start azure virtual machines failed, err: %s", err.Error())
			return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "start azure virtual machines failed")
		}
		pollers = append(pollers, poller)
	}
	for _, poller := range pollers {
		_, err := poller.PollUntilDone(ctx, nil)
		if err != nil {
			log.Errorf(ctx, "poll start azure virtual machines result failed, err: %s", err.Error())
			return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "start azure virtual machines failed")
		}
	}
	return spec.Success()

}

func (be *VmExecutor) stopInstances(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, resourceGroup string, vmNamesArray []string) *spec.Response {
	client, err := CreateVmClient(tenantId, clientId, clientSecret, subscriptionId)
	if err != nil {
		log.Errorf(ctx, "create azure client failed, err: %s", err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create azure client failed")
	}
	pollers := make([]*runtime.Poller[armcompute.VirtualMachinesClientPowerOffResponse], 0)
	for _, vmName := range vmNamesArray {
		poller, err := client.BeginPowerOff(ctx, resourceGroup, vmName, nil)
		if err != nil {
			log.Errorf(ctx, "stop azure virtual machines failed, err: %s", err.Error())
			return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "stop azure virtual machines failed")
		}
		pollers = append(pollers, poller)
	}
	for _, poller := range pollers {
		_, err := poller.PollUntilDone(ctx, nil)
		if err != nil {
			log.Errorf(ctx, "poll stop azure virtual machines result failed, err: %s", err.Error())
			return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "stop azure virtual machines failed")
		}
	}
	return spec.Success()
}

func (be *VmExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}

func CreateVmClient(tenantId, clientId, clientSecret, subscriptionId string) (*armcompute.VirtualMachinesClient, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		// TODO
	}
	options := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Cloud: cloud.AzurePublic,
		},
	}
	return armcompute.NewVirtualMachinesClient(subscriptionId, cred, options)
}

func CreateDiskClient(tenantId, clientId, clientSecret, subscriptionId string) (*armcompute.DisksClient, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		// TODO
	}
	options := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Cloud: cloud.AzurePublic,
		},
	}
	return armcompute.NewDisksClient(subscriptionId, cred, options)
}
