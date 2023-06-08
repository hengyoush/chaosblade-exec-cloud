package azure

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/chaosblade-io/chaosblade-exec-cloud/exec/category"
	"github.com/chaosblade-io/chaosblade-spec-go/log"
	"github.com/chaosblade-io/chaosblade-spec-go/spec"

	"github.com/chaosblade-io/chaosblade-spec-go/util"
)

const DiskBin = "chaos_azure_disk"

type DiskActionSpec struct {
	spec.BaseExpActionCommandSpec
}

func NewDiskActionSpec() spec.ExpActionCommandSpec {
	return &DiskActionSpec{
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
					Desc: "the regionId of azure, like: PUBLIC, CHINA, GOVERMENT. If not provided, get from env AZURE_CLOUD ",
				},
				&spec.ExpFlag{
					Name: "resourceGroup",
					Desc: "the resourceGroup of azure, if not provided, get from env AZURE_RESOURCE_GROUP",
				},
				&spec.ExpFlag{
					Name: "type",
					Desc: "the operation of instances, support attach, detach, etc",
				},
				&spec.ExpFlag{
					Name: "vmname",
					Desc: "the virtual machines name",
				},
				&spec.ExpFlag{
					Name: "diskName",
					Desc: "the disk name",
				},
				&spec.ExpFlag{
					Name: "lun",
					Desc: "the disk lun number",
				},
			},
			ActionExecutor: &DiskExecutor{},
			ActionExample: `
# detach disk 'abc' which virtual machine name is i-x
blade create azure vm --tenantId xxx --clientId xxx --clientSecret xxx --subscriptionId xxx --resourceGroup xxx --type detach --vmname i-x --diskName abc --lun 1

# attach disk 'abc' which virtual machine name is i-x
blade create azure vm --tenantId xxx --clientId xxx --clientSecret xxx --subscriptionId xxx --resourceGroup xxx --type attach --vmname i-x --diskName abc --lun 1`,
			ActionPrograms:   []string{DiskBin},
			ActionCategories: []string{category.Cloud + "_" + category.Azure + "_" + category.Disk},
		},
	}
}

func (*DiskActionSpec) Name() string {
	return "disk"
}

func (*DiskActionSpec) Aliases() []string {
	return []string{}
}
func (*DiskActionSpec) ShortDesc() string {
	return "do some azure disk Operations, like attach, detach"
}

func (b *DiskActionSpec) LongDesc() string {
	if b.ActionLongDesc != "" {
		return b.ActionLongDesc
	}
	return "do some azure disk Operations, like attach, detach"
}

type DiskExecutor struct {
	channel spec.Channel
}

func (*DiskExecutor) Name() string {
	return "disk"
}

func (be *DiskExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	bytes, _ := json.Marshal(*model)
	log.Debugf(ctx, "model: %s", string(bytes))
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
	regionId := model.ActionFlags["regionId"]
	if regionId == "" {
		val, ok := os.LookupEnv("AZURE_CLOUD")
		if !ok {
			log.Errorf(ctx, "could not get AZURE_CLOUD from env or parameter!")
			return spec.ResponseFailWithFlags(spec.ParameterLess, "regionId")
		}
		regionId = val
	}
	vmName := model.ActionFlags["vmname"]
	if vmName == "" {
		log.Errorf(ctx, "vmname is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "vmname")
	}
	diskName := model.ActionFlags["diskName"]
	if diskName == "" {
		log.Errorf(ctx, "diskName is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "diskName")
	}
	lun, err := strconv.Atoi(model.ActionFlags["diskName"])
	if err != nil {
		log.Errorf(ctx, "lun is required and must be valid integer")
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "lun")
	}
	operationType := model.ActionFlags["type"]
	if operationType == "" {
		log.Errorf(ctx, "operationType is required!")
		return spec.ResponseFailWithFlags(spec.ParameterLess, "operationType")
	}
	isAttached, err := isDiskAttached(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName, lun)
	if err != nil {
		return spec.ResponseFailWithFlags(spec.ParameterRequestFailed, "get disks status failed")
	}
	if (isAttached && operationType == "attach") || (!isAttached && operationType == "detach") {
		return be.stop(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, operationType, resourceGroup, vmName, diskName, lun)
	}

	return be.start(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, operationType, resourceGroup, vmName, diskName, lun)
}

func isDiskAttached(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName string, lun int) (bool, error) {
	client, err := CreateVmClient(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId)
	if err != nil {
		log.Errorf(ctx, "create azure client failed, err: %s", err.Error())
		return false, err
	}
	vmResponse, err := client.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		log.Errorf(ctx, "get vm info failed, err: %s", err.Error())
		return false, err
	}
	for _, dataDisk := range vmResponse.VirtualMachine.Properties.StorageProfile.DataDisks {
		if *dataDisk.Lun == int32(lun) && *dataDisk.Name == diskName {
			return true, nil
		}
	}
	return false, nil
}

func (be *DiskExecutor) start(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId, operationType, resourceGroup, vmName, diskName string, lun int) *spec.Response {
	switch operationType {
	case "attach":
		return be.attachDisk(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName, lun)
	case "detach":
		return be.detachDisk(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName, lun)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support attach detach)")
	}
}

func (be *DiskExecutor) stop(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId, operationType, resourceGroup, vmName, diskName string, lun int) *spec.Response {
	switch operationType {
	case "attach":
		return be.detachDisk(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName, lun)
	case "detach":
		return be.attachDisk(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName, lun)
	default:
		return spec.ResponseFailWithFlags(spec.ParameterInvalid, "type is not support(support attach detach)")
	}
}

func (be *DiskExecutor) detachDisk(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName string, lun int) *spec.Response {
	client, err := CreateVmClient(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId)
	if err != nil {
		log.Errorf(ctx, "create azure client failed, err: %s", err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create azure client failed")
	}
	log.Debugf(ctx, "create azure vm client success")
	response, err := client.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		log.Errorf(ctx, "get virtual machine's info failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "get virtual machine's info failed")
	}
	vm := response.VirtualMachine
	diskIndex := -1
	for index, disk := range vm.Properties.StorageProfile.DataDisks {
		if *disk.Lun == int32(lun) && *disk.Name == diskName {
			diskIndex = index
			break
		}
	}
	if diskIndex == -1 {
		log.Errorf(ctx, "try to detach but the disk %s is not attched to the virtual machine %s!", diskName, vmName)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "try to detach but the disk is not attched to the virtual machine")
	}
	newDiskList := append(vm.Properties.StorageProfile.DataDisks[:diskIndex], vm.Properties.StorageProfile.DataDisks[diskIndex+1:]...)
	vm.Properties.StorageProfile.DataDisks = newDiskList
	parameters := armcompute.VirtualMachineUpdate{Identity: vm.Identity, Plan: vm.Plan, Tags: vm.Tags, Zones: vm.Zones, Properties: vm.Properties}
	poller, err := client.BeginUpdate(ctx, resourceGroup, vmName, parameters, nil)
	if err != nil {
		log.Errorf(ctx, "update virtual machine's disk configuration failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "update virtual machine's disk configuration failed")
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Errorf(ctx, "get update virtual machine's disk configuration result failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "get update virtual machine's disk configuration result failed")
	}
	return spec.Success()
}

func (be *DiskExecutor) attachDisk(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId, resourceGroup, vmName, diskName string, lun int) *spec.Response {
	vmClient, err := CreateVmClient(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId)
	if err != nil {
		log.Errorf(ctx, "create azure vm client failed, err: %s", err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create azure client failed")
	}
	diskClient, err := CreateDiskClient(ctx, tenantId, clientId, clientSecret, subscriptionId, regionId)
	if err != nil {
		log.Errorf(ctx, "create azure disk client failed, err: %s", err.Error())
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "create azure client failed")
	}
	log.Debugf(ctx, "create azure vm client success")
	vmResponse, err := vmClient.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		log.Errorf(ctx, "get virtual machine's info failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "get virtual machine's info failed")
	}
	vm := vmResponse.VirtualMachine
	diskResponse, err := diskClient.Get(ctx, resourceGroup, diskName, nil)
	if err != nil {
		log.Errorf(ctx, "get disk id failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "get disk id failed")
	}

	newDiskList := append(vm.Properties.StorageProfile.DataDisks, &armcompute.DataDisk{
		CreateOption: &armcompute.PossibleDiskCreateOptionTypesValues()[0],
		Lun:          toPtr(int32(lun)),
		Name:         &diskName,
		ManagedDisk: &armcompute.ManagedDiskParameters{
			ID: diskResponse.Disk.ID,
		},
	})
	vm.Properties.StorageProfile.DataDisks = newDiskList
	parameters := armcompute.VirtualMachineUpdate{Identity: vm.Identity, Plan: vm.Plan, Tags: vm.Tags, Zones: vm.Zones, Properties: vm.Properties}
	poller, err := vmClient.BeginUpdate(ctx, resourceGroup, vmName, parameters, nil)
	if err != nil {
		log.Errorf(ctx, "update virtual machine's disk configuration failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "update virtual machine's disk configuration failed")
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Errorf(ctx, "get update virtual machine's disk configuration result failed, err: %s", err)
		return spec.ResponseFailWithFlags(spec.ContainerInContextNotFound, "get update virtual machine's disk configuration result failed")
	}
	return spec.Success()
}

func CreateDiskClient(ctx context.Context, tenantId, clientId, clientSecret, subscriptionId, regionId string) (*armcompute.DisksClient, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		log.Errorf(ctx, "create azure disk client failed, err: %s", err.Error())
	}
	options := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Cloud: AzureCloud(regionId),
		},
	}

	return armcompute.NewDisksClient(subscriptionId, cred, options)
}

func toPtr(i int32) *int32 {
	return &i
}

func (be *DiskExecutor) SetChannel(channel spec.Channel) {
	be.channel = channel
}
