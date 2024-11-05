package shellcli

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/xml"
	"fmt"
	"freyja/internal"
	"github.com/digitalocean/go-libvirt"
	"github.com/dypflying/go-qcow2lib/qcow2"
	"github.com/google/uuid"
	"github.com/kdomanski/iso9660"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var configurationPath string

const BackingImageFormat string = "qcow2"

const BackingImageFileSuffix string = "-overlay-image." + BackingImageFormat

const RootImageFileSuffix string = "-root-image." + BackingImageFormat

const ISOCloudInitFileSuffix string = "-cloud-init.iso"

const XMLMachineDescriptionSuffix string = "-libvirt-conf.xml"

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// build config from path
		var configurationData internal.ConfigurationData
		if err := configurationData.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err)
			os.Exit(1)
		}
		// build cloud init config file
		var cloudInitData internal.CloudInitUserData
		for _, machine := range configurationData.Machines {
			if err := cloudInitData.Build(&machine); err != nil {
				Logger.Error("cannot build cloud init specs from machine configuration", "configuration", machine, "reason", err)
				os.Exit(1)
			}
			// create machine directory
			machineDirPath, err := createMachineDir(&machine)
			if err != nil {
				Logger.Error("cannot create machine workspace directory", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			// create cloud init metadata and user data files
			if err := internal.GenerateCloudInitConfigs(&machine, machineDirPath); err != nil {
				Logger.Error("cannot create machine cloud init configurations", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			// create cloud-init iso file
			cloudInitIsoFile, err := createCloudInitIso(&machine, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine ISO image file", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}

			// copy root image to the machine dir
			rootImagePath := filepath.Join(machineDirPath, machine.Hostname+RootImageFileSuffix)
			if err := internal.CopyFile(machine.Image, rootImagePath, 0700); err != nil {
				Logger.Error("Cannot copy machine root image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
			// use 'qemu-img info' to verify it
			overlayFile, err := createOverlayImage(&machine, rootImagePath, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine overlay image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// TODO :
			//   cloud init not taken into account at boot.
			//   succeeded one time by rebooting but never again.
			//   - try to check how to consider cloud init in boot phase
			//   - try to adapt the cloud init file maybe and recreate the vm raw image and cloud init iso image ?
			//   - try simply with 'user' and 'password' in the cloud init user data file
			//   - try to manually create the cloud-init iso with genisoimage and create the domain using virsh
			// TODO :
			//  fix xml conf errors in input :
			//   - vcpu needs a attr 'placement="static"'
			//   - no need to generate an id for the root element 'domain', libvirt does it
			//   - add an element 'description' equal to the domain name
			//   - investigate on :
			//        <on_poweroff>destroy</on_poweroff>
			//  	  <on_reboot>restart</on_reboot>
			//  	  <on_crash>destroy</on_crash>
			//   - investigate on :
			//      VFs do not get a fixed MAC address; it changes every time the host reboots.
			//      When adding network devices the “traditional” way with hostdev, it would require
			//      to reconfigure the VM Guest's network device after each reboot of the host,
			//      because of the MAC address change. To avoid this kind of problem, libvirt
			//      introduced the hostdev value, which sets up network-specific data before assigning the device.
			// example of command : https://sumit-ghosh.com/posts/create-vm-using-libvirt-cloud-images-cloud-init/
			// example of script :
			//virt-install \
			//    --connect=qemu:///system \
			//    --import \
			//    --name "${HOSTNAME}" \
			//    --memory "${MEMORY_SIZE}" \
			//    --vcpus "${VCPUS_NUMBER}" \
			//    --cpu host,+vmx \
			//    --metadata description="${HOSTNAME}" \
			//    --os-variant "${OS_VARIANT}" \
			//    --disk "path=${instanciated_image},readonly=false" \
			//    --disk "${cloud_init_image},device=cdrom" \
			//    --hvm \
			//    --graphics none \
			//    --noautoconsole \
			//    --network network=ctrl-plane,mac=52:54:02:29:e3:cc
			xmlMachineDescription, err := createLibvirtDomainXMLDescription(&machine, overlayFile, cloudInitIsoFile)
			if err != nil {
				Logger.Error("cannot create the libvirt domain XML description from machine configuration", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// dump description in machine dir (useful for debug)
			xmlMachineDescriptionPath := filepath.Join(machineDirPath, machine.Hostname+XMLMachineDescriptionSuffix)
			if err := os.WriteFile(xmlMachineDescriptionPath, xmlMachineDescription, 0660); err != nil {
				Logger.Error("cannot write the libvirt domain XML description in config dir", "machine", machine.Hostname, "path", xmlMachineDescriptionPath, "reason", err)
				os.Exit(1)
			}

			// create the machine in libvirt
			_, err = LibvirtConnexion.DomainCreateXML(string(xmlMachineDescription), libvirt.DomainNone)
			if err != nil {
				Logger.Error("cannot create the machine from libvirt domain XML description", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			log.Print("OK")
		}

	},
}

func init() {
	machineCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := machineCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
}

// createMachineDir returns the created dir, or an error
func createMachineDir(machine *internal.ConfigurationMachine) (string, error) {
	machineDirPath := filepath.Join(FreyjaMachinesWorkspaceDir, machine.Hostname)
	if _, err := os.Stat(machineDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(machineDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return machineDirPath, nil
}

// Implementation using kdomanski/iso9660 : https://github.com/kdomanski/iso9660
// it is also used in terraform for proxmox : https://github.com/Telmate/terraform-provider-proxmox/blob/186ec3f23bf4a62fcad35f6292fa1350b8e1183b/proxmox/resource_cloud_init_disk.go#L77-L122
func createCloudInitIso(machine *internal.ConfigurationMachine, machineDir string) (string, error) {
	isoWriter, err := iso9660.NewWriter()
	if err != nil {
		return "", fmt.Errorf("cannot create the iso9660 writer for the machine '%s' : %w", machine.Hostname, err)
	}
	// add cloud init metadata
	metadataPath := filepath.Join(machineDir, internal.GetCloudInitMetadataFilename(machine.Hostname))
	fm, err := os.Open(metadataPath)
	if err != nil {
		return "", fmt.Errorf("cannot open the cloud init metadata file in '%s' : %w", metadataPath, err)
	}
	defer fm.Close()
	if err = isoWriter.AddFile(fm, "meta-data"); err != nil {
		return "", fmt.Errorf("cannot add the cloud init metadata file in ISO for the machine '%s' : %w", machine.Hostname, err)
	}
	// add cloud init user data
	userdataPath := filepath.Join(machineDir, internal.GetCloudInitUserDataFilename(machine.Hostname))
	fu, err := os.Open(userdataPath)
	if err != nil {
		return "", fmt.Errorf("cannot open the cloud init user data file in '%s' : %w", userdataPath, err)
	}
	defer fu.Close()
	if err = isoWriter.AddFile(fu, "user-data"); err != nil {
		return "", fmt.Errorf("cannot add the cloud init user data file in ISO for the machine '%s' : %w", machine.Hostname, err)
	}
	// write iso on filesystem
	isoOutputPath := filepath.Join(machineDir, fmt.Sprintf("%s%s", machine.Hostname, ISOCloudInitFileSuffix))
	outputFile, err := os.OpenFile(isoOutputPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", fmt.Errorf("cannot open the ISO image file path '%s' : %w", isoOutputPath, err)
	}
	// calculate the iso file ID
	h := sha256.New()
	h.Write([]byte(isoOutputPath))
	isoID := h.Sum(nil)
	if err = isoWriter.WriteTo(outputFile, string(isoID)); err != nil {
		return "", fmt.Errorf("cannot write the ISO image file in '%s' : %w", isoOutputPath, err)
	}
	if err = outputFile.Close(); err != nil {
		return "", err
	}

	return isoOutputPath, nil
}

func createOverlayImage(machine *internal.ConfigurationMachine, rootImagePath string, machineDir string) (string, error) {
	// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
	opts := make(map[string]any)
	backingFile, err := filepath.Abs(rootImagePath)
	if err != nil {
		return "", fmt.Errorf("cannot read base image file '%s' : %w", machine.Image, err)
	}
	overlayFile := filepath.Join(machineDir, machine.Hostname+BackingImageFileSuffix)
	opts[qcow2.OPT_SIZE] = machine.Storage << 30 //qcow2 file's size is 1g
	opts[qcow2.OPT_FMT] = BackingImageFormat     //qcow2 format
	opts[qcow2.OPT_SUBCLUSTER] = true            //enable sub-cluster
	opts[qcow2.OPT_BACKING] = backingFile
	if err := qcow2.Blk_Create(overlayFile, opts); err != nil {
		return "", fmt.Errorf("failed to create overlay qcow2 file '%s' : %w", overlayFile, err)
	}

	return overlayFile, nil
}

func createLibvirtDomainXMLDescription(cm *internal.ConfigurationMachine, overlayFile string, cloudInitIsoFile string) ([]byte, error) {
	// TODO: how to generate the unique id ?
	//  - use '1', '2', ... and store the existing ids in metadata to make sure to not reuse an existing one ?
	//  - use a uuid ?
	//  - use a hash of the hostname ? -> choice 1
	mName := cm.Hostname
	// ID is the fnv hash of the machine hostname
	h := sha256.New()
	_, err := h.Write([]byte(mName))
	if err != nil {
		return nil, fmt.Errorf("cannot generate a hash based on the machine hostname '%s': %w", cm.Hostname, err)
	}
	sum := h.Sum(nil)
	mID := b64.StdEncoding.EncodeToString(sum)[:16]
	// UUID is generated from the ID
	mUUIDRaw, err := uuid.FromBytes(sum[:16])
	if err != nil {
		return nil, fmt.Errorf("cannot create machine UUID based on its ID '%s': %w", string(mID), err)
	}
	mUUID := fmt.Sprintf("%v", mUUIDRaw)
	// memory
	mMemoryValue := uint64(internal.MiBToKiB(uint64(cm.Memory)))
	mMemoryUnit := string(internal.KiBMemoryUnit)
	mMemory := internal.XMLDomainDescriptionMemory{
		Unit:  mMemoryUnit,
		Value: mMemoryValue,
	}
	// cpu
	mVcpu := cm.Vcpu
	// OS description
	// TODO : boot element ? like <boot dev='hd'/>
	Arch := string(internal.X86OSArch)
	mOSType := internal.XMLDomainDescriptionOSType{
		Type: internal.DefaultOsType,
		Arch: Arch,
	}
	mOS := internal.XMLDomainDescriptionOS{
		Type: &mOSType,
	}
	// metadata
	// TODO : find a way to get the proper os type from golang
	//  for now, no metadata for os info part

	// live image device
	//	    <disk device="disk" type="file">
	//	        <driver name="qemu" type="qcow2"/>
	//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_vdisk.debian12" index="2"/>
	//	        <backingStore index="3" type="file">
	//	            <format type="qcow2"/>
	//	            <source file="/home/kaio/Images/debian12"/>
	//	            <backingStore/>
	//	        </backingStore>
	//	        <target bus="virtio" dev="vda"/>
	//	        <alias name="virtio-disk0"/>
	//	        <address bus="0x04" domain="0x0000" function="0x0" slot="0x00" type="pci"/>
	//	    </disk>
	// TODO :
	//   - check if image source path can be common to multiple machines (from the moment the overlays are different for the machines)
	//   - if not, copy the source image file to the working dir of the machine
	liveImageDevice := internal.XMLDomainDescriptionDevicesDisk{
		Device: string(internal.DiskDeviceType),
		Type:   string(internal.FileDeviceDiskType),
		Driver: &internal.XMLDomainDescriptionDevicesDiskDriver{
			Name: string(internal.QemuDeviceDiskDriverName),
			Type: string(internal.QcowDeviceDiskDriverType),
		},
		Source: &internal.XMLDomainDescriptionDevicesDiskSource{
			//File: cm.Image,
			File: overlayFile,
		},
		BackingStore: &internal.XMLDomainDescriptionDevicesDiskBackingStore{
			Type: string(internal.FileDeviceDiskType),
			Format: &internal.XMLDomainDescriptionDevicesDiskBackingStoreFormat{
				Type: string(internal.QcowDeviceDiskDriverType),
			},
			Source: &internal.XMLDomainDescriptionDevicesDiskBackingStoreSource{
				//File: overlayFile,
				File: cm.Image,
			},
		},
		Target: &internal.XMLDomainDescriptionDevicesDiskTarget{
			Bus:    string(internal.VirtioDeviceDiskTargetBus),
			Device: string(internal.VdaDeviceDiskTargetDev),
		},
	}

	// cloud init raw iso device
	//	    <disk device="cdrom" type="file">
	//	        <driver name="qemu" type="raw"/>
	//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_cloud_init.iso" index="1"/>
	//	        <backingStore/>
	//	        <target bus="sata" dev="sda"/>
	//	        <readonly/>
	//	        <alias name="sata0-0-0"/>
	//	        <address bus="0" controller="0" target="0" type="drive" unit="0"/>
	//	    </disk>
	cloudInitIsoDevice := internal.XMLDomainDescriptionDevicesDisk{
		Device: string(internal.CdromDeviceType),
		Type:   string(internal.FileDeviceDiskType),
		Driver: &internal.XMLDomainDescriptionDevicesDiskDriver{
			Name: string(internal.QemuDeviceDiskDriverName),
			Type: string(internal.RawDeviceDiskDriverType),
		},

		Source: &internal.XMLDomainDescriptionDevicesDiskSource{
			File: cloudInitIsoFile,
		},
		Target: &internal.XMLDomainDescriptionDevicesDiskTarget{
			Bus:    string(internal.SataDeviceDiskTargetBus),
			Device: string(internal.SdaDeviceDiskTargetDev),
		},
	}

	// network device
	//	    <interface type="network">
	//	        <mac address="52:54:00:25:77:0d"/>
	//	        <source bridge="virbr0" network="default" portid="5b2b65a8-8c46-4109-9117-38e4bbef3cd6"/>
	//	        <target dev="vnet0"/>
	//	        <model type="virtio"/>
	//	        <alias name="net0"/>
	//	        <address bus="0x01" domain="0x0000" function="0x0" slot="0x00" type="pci"/>
	//	    </interface>
	var bridgeInterfaceDevices []internal.XMLDomainDescriptionDevicesInterface
	if len(cm.Networks) > 0 {
		// user defined networks
		bridgeInterfaceDevices = make([]internal.XMLDomainDescriptionDevicesInterface, len(cm.Networks))
		for _, network := range cm.Networks {
			// TODO :
			//  - investigate how to provide a target if the user configure an interface name on the host
			//  - configure interface model
			bridgeInterfaceDevice := internal.XMLDomainDescriptionDevicesInterface{
				Type: string(internal.NetworkDeviceInterfaceType),
				Mac: &internal.XMLDomainDescriptionDevicesInterfaceMac{
					Address: network.Mac,
				},
				Source: &internal.XMLDomainDescriptionDevicesInterfaceSource{
					Bridge:  internal.DefaultInterfaceSourceBridge,
					Network: network.Name,
				},
				Target: nil,
				//Target: internal.XMLDomainDescriptionDevicesInterfaceTarget{}, // provide if user conf specifies a host interface
				Model: &internal.XMLDomainDescriptionDevicesInterfaceModel{
					Type: internal.DefaultInterfaceModelType,
				},
			}
			bridgeInterfaceDevices = append(bridgeInterfaceDevices, bridgeInterfaceDevice)
		}
	} else {
		// if no network define in user conf input, stick with the default one
		bridgeInterfaceDevices = []internal.XMLDomainDescriptionDevicesInterface{internal.DefaultDeviceInterface}
	}

	// console and serial device for graphical debug
	// they must target the same serial port
	//	        <serial type='pty'>
	//	          <target type='isa-serial' port='0'>
	//	            <model name='isa-serial'/>
	//	          </target>
	//	        </serial>
	//	        <console type='pty'>
	//	          <target type='serial' port='0'/>
	//	        </console>
	targetSerialPort := "0"
	serialDevice := internal.XMLDomainDescriptionDevicesSerial{
		Type: string(internal.PtyDeviceSerialType),
		Target: &internal.XMLDomainDescriptionDevicesSerialTarget{
			Type: string(internal.IsaDeviceSerialTargetType),
			Port: targetSerialPort,
			Model: &internal.XMLDomainDescriptionDevicesSerialTargetModel{
				Name: string(internal.IsaDeviceSerialTargetModelName),
			},
		},
	}
	consoleDevice := internal.XMLDomainDescriptionDevicesConsole{
		Type: string(internal.PtyDeviceConsoleType),
		Target: &internal.XMLDomainDescriptionDevicesConsoleTarget{
			Type: string(internal.SerialDeviceConsoleTargetType),
			Port: targetSerialPort,
		},
	}

	// xml
	deviceDisks := []internal.XMLDomainDescriptionDevicesDisk{liveImageDevice, cloudInitIsoDevice}
	xmlDescription := internal.XMLDomainDescription{
		Type:   internal.DefaultDomainType,
		ID:     mID,
		Name:   mName,
		UUID:   mUUID,
		Vcpu:   mVcpu,
		Memory: &mMemory,
		OS:     &mOS,
		Devices: &internal.XMLDomainDescriptionDevices{
			Emulator:   string(internal.QemuX86DevicesEmulator),
			Disks:      deviceDisks,
			Interfaces: bridgeInterfaceDevices,
			Serial:     []internal.XMLDomainDescriptionDevicesSerial{serialDevice},
			Console:    []internal.XMLDomainDescriptionDevicesConsole{consoleDevice},
		},
	}
	return xml.Marshal(xmlDescription)
}
