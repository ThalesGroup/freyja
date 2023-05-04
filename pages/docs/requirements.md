---
sidebar_label: 'Requirements'
sidebar_position: 2
---

# Requirements

:::danger Mandatory
You must install first on your hypervisor :
* **Qemu-KVM** & **Libvirt**
* **Python >= 3.9**
:::

:::caution Provision
For *classic* OS provisioning, you must install on your hypervisor:
* **Cloud-Init**

It concerns Debian-based OS, RedHat-based OS, Alpine, ArchLinux, Gentoo, etc...  
Check the [complete list of supported distributions](https://cloudinit.readthedocs.io/en/latest/reference/availability.html#distributions).
:::

:::info Immutable OS
If you plan to install immutable OS only, you don't need Cloud-Init.  
You must provision the OS using **[Ignition files](https://coreos.github.io/ignition/)**.
:::