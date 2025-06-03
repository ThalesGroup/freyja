# troubleshooting

## La vm n'accroche que le premier réseau libvirt configuré

**ma question**

J'ai un souci avec une VM créée en utilisant virsh et une description xml : une des 2 interfaces définies dans le xml ne possède pas d'IP dans la VM. Cette interface utilise un réseau que j'ai créé dans libvirt et qui s'appelle dataplane.

Voici la description xml de cette VM :

```xml
<domain type="kvm" id="2">
  <name>vm1</name>
  <uuid>7b11e3d1-b4ef-77af-7e81-aa2aba7af47f</uuid>
  <memory unit="KiB">4194304</memory>
  <currentMemory unit="KiB">4194304</currentMemory>
  <vcpu placement="static">1</vcpu>
  <resource>
    <partition>/machine</partition>
  </resource>
  <os>
    <type arch="x86_64" machine="pc-i440fx-7.2">hvm</type>
    <boot dev="hd"/>
  </os>
  <cpu mode="custom" match="exact" check="full">
    <model fallback="forbid">qemu64</model>
    <feature policy="require" name="x2apic"/>
    <feature policy="require" name="hypervisor"/>
    <feature policy="require" name="lahf_lm"/>
    <feature policy="disable" name="svm"/>
  </cpu>
  <clock offset="utc"/>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="/home/kaio/.freyja/machines/vm1/overlay-image.qcow2" index="2"/>
      <backingStore type="file" index="3">
        <format type="qcow2"/>
        <source file="/tmp/debian-12-generic-amd64.qcow2"/>
        <backingStore/>/>
      <target dev="hda" bus="ide"/>
      <alias name="ide0-0-0"/>
      <address type="drive" controller="0" bus="0" target="0" unit="0"/>
    </disk>
    <disk type="file" device="cdrom">
      <driver name="qemu" type="raw"/>
      <source file="/home/kaio/.freyja/machines/vm1/cloud-init.iso" index="1"/>
      <backingStore/>
      <target dev="hdb" bus="ide"/>
      <readonly/>
      <alias name="ide0-0-1"/>
      <address type="drive" controller="0" bus="0" target="0" unit="1"/>
    </disk>
    <controller type="usb" index="0" model="piix3-uhci">
      <alias name="usb"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x01" function="0x2"/>
    </controller>
    <controller type="pci" index="0" model="pci-root">
      <alias name="pci.0"/>
    </controller>
    <controller type="ide" index="0">
      <alias name="ide"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x01" function="0x1"/>
    </controller>
    <interface type="network">
      <mac address="52:54:00:a2:48:c9"/>
      <source network="ctrlplane" portid="c2c46318-f609-4200-9d51-07556d5cb747" bridge="virbr1"/>
      <target dev="vnet0"/>
      <model type="rtl8139"/>
      <alias name="net0"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x0"/>
    </interface>
    <interface type="network">
      <mac address="52:54:00:f0:ca:2a"/>
      <source network="dataplane" portid="29865fda-3535-4625-b1df-3d3bf82bebf4" bridge="virbr2"/>
      <target dev="vnet1"/>
      <model type="rtl8139"/>
      <alias name="net1"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x03" function="0x0"/>
    </interface>
    <serial type="pty">
      <source path="/dev/pts/3"/>
      <target type="isa-serial" port="0">
        <model name="isa-serial"/>
      </target>
      <alias name="serial0"/>
    </serial>
    <console type="pty" tty="/dev/pts/3">
      <source path="/dev/pts/3"/>
      <target type="serial" port="0"/>
      <alias name="serial0"/>
    </console>
    <input type="mouse" bus="ps2">
      <alias name="input0"/>
    </input>
    <input type="keyboard" bus="ps2">
      <alias name="input1"/>
    </input>
    <audio id="1" type="none"/>
    <memballoon model="virtio">
      <alias name="balloon0"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x04" function="0x0"/>
    </memballoon>
  </devices>
  <seclabel type="dynamic" model="apparmor" relabel="yes">
    <label>libvirt-7b11e3d1-b4ef-77af-7e81-aa2aba7af47f</label>
    <imagelabel>libvirt-7b11e3d1-b4ef-77af-7e81-aa2aba7af47f</imagelabel>
  </seclabel>
  <seclabel type="dynamic" model="dac" relabel="yes">
    <label>+1000:+1000</label>
    <imagelabel>+1000:+1000</imagelabel>
  </seclabel>
</domain>
```

On voit qu'elle utilise 2 réseaux : ctrplane et dataplane.
Ces 2 réseaux ont été créés avec `virsh net-define` puis `virsh net-create` en utilisant des fichiers xml. Voici l'exemple pour dataplane :

```xml
<network>
  <name>dataplane</name>
  <ip address="192.168.124.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.124.2" end="192.168.124.254"/>
    </dhcp>
  </ip>
</network>
```

Puis, j'ai créé vm1 dans libvirt avec `virsh define` puis `virsh create` en utilisant la description xml que j'ai donnée plus haut.

La machine démarre bien puis en m'y connectant j'inspecte les addresses depuis l'intérieur de la machine avec `ip r` et j'obtiens :

```sh
192.168.123.0/24 dev enp0s2 proto kernel scope link src 192.168.123.233 metric 100 
192.168.123.1 dev enp0s2 proto dhcp scope link src 192.168.123.233 metric 100
```

On voit que ce n'est pas dans le range dhcp de dataplane. Ici, il s'agit en fait du range dhcp de ctrlplane, donc cette vm n'a pas d'IP qui utilise le réseau dataplane précédemment créé.

En inspectant dans la machine avec `ip a` :

```sh
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host noprefixroute 
       valid_lft forever preferred_lft forever
2: enp0s2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 52:54:00:a2:48:c9 brd ff:ff:ff:ff:ff:ff
    inet 192.168.123.233/24 metric 100 brd 192.168.123.255 scope global dynamic enp0s2
       valid_lft 3585sec preferred_lft 3585sec
    inet6 fe80::5054:ff:fea2:48c9/64 scope link 
       valid_lft forever preferred_lft forever
3: enp0s3: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 52:54:00:f0:ca:2a brd ff:ff:ff:ff:ff:ff
```

On voit en effet que l'IP 192.168.123.233 utilise l'interface enp0s2 mais que l'interface enp0s3 ne comporte aucune IP.

En inspectant de nouveau virsh depuis l'extérieur de la machine avec `virsh domiflist vm1` j'obtiens :

```sh
Interface Type Source Model MAC
vnet0 network ctrlplane rtl8139 52:54:00:a2:48:c9
vnet1 network dataplane rtl8139 52:54:00:f0:ca:2a
```

On voit bien que l'adresse mac 52:54:00:f0:ca:2a qui ne comporte aucune IP dans la VM vm1, rattachée à l'interface enp0s3, est bien celle de dataplane.

Pourquoi n'ai-je pas d'IP pour cette interface ? Pourquoi seulement ctrlplane comporte une IP pour la vm ? Comment obtenir une IP pour dataplane ? Tout en conservant la méthode de création des réseaux et des machines libvirt en utilisant des configurations xml. 

















**RÉPONSE CHATGPT**

!!! ESSAYER LA MÉTHODE CLOUDINIT !!!

Merci pour la clarté, c’est un diagnostic bien structuré, et toutes les infos clés sont là. Voici une analyse détaillée de la situation et les pistes de résolution.

---

## 🔎 **Résumé du problème**
- Deux interfaces `virsh` : `ctrlplane` (vnet0) et `dataplane` (vnet1)
- **Seule** l’interface `ctrlplane` reçoit une IP via DHCP dans la VM.
- `dataplane` est bien visible dans la VM (enp0s3), **mais elle est down** et **n’a pas d’adresse IP**.

---

## ✅ Ce qui fonctionne

- Le réseau **ctrlplane** fonctionne : DHCP attribue bien une IP (`192.168.123.x`) sur l'interface `enp0s2`.
- L'interface réseau `dataplane` est bien **attachée au domaine** (virsh confirme le mapping MAC / nom).
- L’interface `enp0s3` (dataplane) est bien présente **dans la VM**, mais elle est `DOWN` et sans IP.

---

## 📍 Analyse

Voici les causes **possibles** à ton problème :

### 1. 💡 **Pas de service DHCP réellement actif sur `dataplane`**
Ta définition XML de `dataplane` semble correcte :

```xml
<network>
  <name>dataplane</name>
  <ip address="192.168.124.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.124.2" end="192.168.124.254"/>
    </dhcp>
  </ip>
</network>
```

Mais : le fait d'utiliser `virsh net-define` + `virsh net-create` **sans `<bridge>` explicite** peut créer un réseau **isolé** ou **non activé correctement**. Il est possible que le service `dnsmasq` pour `dataplane` **ne tourne pas**.

Tu peux vérifier ça avec :
```sh
ps aux | grep dnsmasq
```

Tu dois y voir un processus lié à `dataplane`, typiquement avec un fichier de conf dans `/var/lib/libvirt/dnsmasq/dataplane.*`.

Autre vérification :
```sh
virsh net-info dataplane
```

Tu dois y voir :
- `Active: yes`
- `Autostart: yes`
- `Persistent: yes`
- Et idéalement un bridge `virbr2` bien actif dans `ip a`.

### 2. 🔌 **Interface `enp0s3` non montée dans la VM (cloud-init ou netplan)**

!!! ESSAYER CETTE METHODE !!!

Même si l’interface est vue au boot, **elle n'est pas activée**. Typiquement, si tu utilises **cloud-init**, **netplan** ou même un fichier `/etc/network/interfaces`, il faut que l'interface secondaire soit explicitement configurée.

💡 Si `enp0s3` est inconnue de la config réseau dans la VM, alors elle ne montera pas (même si `dnsmasq` est fonctionnel côté libvirt).

---

## ✅ Étapes de Dépannage

### 1. **Vérifie que le DHCP est bien actif pour `dataplane`**

```sh
sudo virsh net-dumpxml dataplane
```

Tu devrais y voir une section `<bridge name="virbr2">`, et éventuellement `<forward mode="nat"/>` ou `none`.

Puis :
```sh
ip a | grep virbr2
```

Tu dois voir que le bridge existe bien sur l’hôte.

Et enfin :
```sh
ps aux | grep dataplane
```

Tu dois voir un `dnsmasq` lié au réseau `dataplane`. Sinon, le DHCP n'est pas lancé.

---

### 2. **Essaye une attribution manuelle depuis la VM**

Dans la VM, active manuellement l’interface pour voir si une IP est attribuée :

```sh
sudo ip link set enp0s3 up
sudo dhclient enp0s3
```

Et regarde :

```sh
ip a show enp0s3
```

Si tu reçois une IP dans le range `192.168.124.x`, **alors ton bridge/dnsmasq fonctionne**, mais **la VM ne configure pas automatiquement cette interface** au boot.

---

### 3. **Corrige la config réseau dans la VM**

En fonction de ta distro (tu parles de Debian 12), tu utilises probablement `netplan` ou `systemd-networkd`.

🔧 **Si cloud-init est utilisé**, assure-toi que ton `cloud-init.cfg` (ou `network-config`) configure bien les deux interfaces. Sinon, ajoute un fichier netplan comme suit dans la VM :

#### Exemple de config netplan :

```yaml
network:
  version: 2
  ethernets:
    enp0s2:
      dhcp4: true
    enp0s3:
      dhcp4: true
```

Puis :

```sh
sudo netplan apply
```

Ou :

```sh
sudo systemctl restart systemd-networkd
```

---

### 4. **Optionnel : Ajoute un `<bridge>` explicite à la définition de ton réseau dataplane**

Si tu veux plus de contrôle :

```xml
<network>
  <name>dataplane</name>
  <bridge name="virbr2"/>
  <ip address="192.168.124.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.124.2" end="192.168.124.254"/>
    </dhcp>
  </ip>
</network>
```

Puis relance :

```sh
virsh net-destroy dataplane
virsh net-undefine dataplane
virsh net-define dataplane.xml
virsh net-start dataplane
virsh net-autostart dataplane
```

---

## 🧩 Autres pistes avancées (si tout le reste échoue)

- Inspecte les logs dans `/var/log/libvirt/qemu/vm1.log` ou `/var/log/syslog` (ou journalctl).
- Utilise `tcpdump -i virbr2` pour voir si les requêtes DHCP sortent.
- Vérifie que le MAC `52:54:00:f0:ca:2a` est bien dans les leases DNSMasq :

```sh
cat /var/lib/libvirt/dnsmasq/dataplane.leases
```

---

## ✅ En résumé

| Étape                        | Action                                                   |
|-----------------------------|-----------------------------------------------------------|
| Vérifie que `dataplane` est actif | `virsh net-info dataplane` + `dnsmasq` présent         |
| Teste manuellement          | `ip link set enp0s3 up` + `dhclient enp0s3`              |
| Vérifie dans la VM          | Ajoute config Netplan pour `enp0s3`                      |
| Redéfinis ton réseau        | Ajoute `<bridge name="virbr2"/>` dans la config XML     |
| Inspecte les logs/le trafic | `journalctl`, `tcpdump`, ou `*.leases`                   |

---

Si tu veux, je peux aussi te générer une config `netplan` complète adaptée à ta VM. Tu veux du statique ou du DHCP pour l’interface `dataplane` ?

##

[FAILED] Failed to start cloud-init…l cloud-init job (pre-networking).

