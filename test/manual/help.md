# Guide manuel

Télécharger une image qcow2.

Utiliser les fichiers `user-data` et `meta-data` pour créer un iso cloud-init :

```sh
genisoimage -output /home/kaio/Projects/thales/freyja/test/manual/debian12-cloud-init.iso -volid cidata -joliet -rock /home/kaio/Projects/thales/freyja/test/manual/user-data /home/kaio/Projects/thales/freyja/test/manual/meta-data
```

Définir la vm dans libvirt :

```sh
virsh define vm-manual.xml
Domain 'debian12-manual' defined from vm-manual.xml
```

Démarrer la machine :

```sh
virsh start debian12-manual
```

## Chatgpt

Pour créer et initialiser une machine virtuelle Debian 12 sur un système Linux en utilisant qemu-kvm, libvirt, et cloud-init, voici les étapes à suivre :
Étape 1 : Préparer le fichier qcow2 pour le disque principal

Assurez-vous que l’image .qcow2 de Debian 12 est bien dans un emplacement accessible. Pour cet exemple, nous supposerons qu’elle est placée dans /var/lib/libvirt/images/debian12.qcow2.
Étape 2 : Créer les fichiers cloud-init pour la configuration de la VM

Pour initialiser la VM avec cloud-init, vous avez besoin de deux fichiers : user-data et meta-data.

    user-data : contient les informations de configuration pour l'initialisation (comme le nom d'utilisateur et le mot de passe).
    meta-data : contient les informations de base sur la machine virtuelle (nom de l'instance, etc.).

Voici un exemple simple de chaque fichier.

Fichier user-data:

yaml

#cloud-config
users:
- name: debianuser
  sudo: ["ALL=(ALL) NOPASSWD:ALL"]
  groups: sudo
  shell: /bin/bash
  lock_passwd: false
  passwd: "$6$rounds=4096$AsyZhB.cPj$tOXuF11r2quQ.Y2IbHuZcgk8Qa8hEr81CjeZTS7y41GrL54/7k1Lo9Zm8Jh0kE9ay/xhsBZT6XBbVi01KtQjz."
chpasswd:
  expire: false
ssh_pwauth: true

Dans cet exemple, l'utilisateur s'appelle debianuser et le mot de passe est configuré via passwd.

Fichier meta-data:

yaml

instance-id: debian12
local-hostname: debian-vm

    Note : L'instance ID et le hostname sont arbitraires. Vous pouvez les personnaliser selon vos besoins.

Étape 3 : Créer un disque ISO cloud-init avec les fichiers de configuration

Ces fichiers doivent être intégrés dans un disque ISO que la VM pourra lire lors du démarrage.

bash

# Créer un dossier pour stocker les fichiers
mkdir -p /var/lib/libvirt/images/cloud-init/

# Copier les fichiers user-data et meta-data dans ce dossier
cp user-data meta-data /var/lib/libvirt/images/cloud-init/

# Créer l'ISO de cloud-init
genisoimage -output /var/lib/libvirt/images/debian12-cloud-init.iso -volid cidata -joliet -rock /var/lib/libvirt/images/cloud-init/user-data /var/lib/libvirt/images/cloud-init/meta-data

Étape 4 : Créer le fichier XML de description de la VM

Le fichier XML pour la VM définit ses caractéristiques matérielles et l'emplacement des disques.

Voici un exemple de fichier XML pour votre VM Debian :

xml

<domain type="kvm">
  <name>debian12</name>
  <memory unit="MiB">2048</memory>
  <vcpu placement="static">2</vcpu>
  <os>
    <type arch="x86_64" machine="pc-i440fx-2.9">hvm</type>
  </os>
  <devices>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="/var/lib/libvirt/images/debian12.qcow2"/>
      <target dev="vda" bus="virtio"/>
    </disk>
    <disk type="file" device="cdrom">
      <driver name="qemu" type="raw"/>
      <source file="/var/lib/libvirt/images/debian12-cloud-init.iso"/>
      <target dev="vdb" bus="virtio"/>
      <readonly/>
    </disk>
    <interface type="network">
      <mac address="52:54:00:00:00:01"/>
      <source network="default"/>
      <model type="virtio"/>
    </interface>
    <console type="pty"/>
    <graphics type="vnc" port="-1" autoport="yes"/>
  </devices>
</domain>

    Note :

        Modifiez le chemin de l’image de disque et de l’ISO cloud-init selon leur emplacement réel sur votre système.
        Ajustez la configuration des ressources (mémoire, CPU) selon vos besoins.

Étape 5 : Définir et démarrer la machine virtuelle avec virsh

Maintenant, nous allons utiliser virsh pour définir et démarrer la machine virtuelle.

    Définir la VM avec le fichier XML :

    bash

sudo virsh define /path/to/debian12.xml

Démarrer la VM :

bash

sudo virsh start debian12

Vérifier le statut de la VM (optionnel) :

bash

    sudo virsh list --all

La machine virtuelle Debian 12 devrait maintenant démarrer et être initialisée avec les configurations fournies par cloud-init.
Étape 6 : Accéder à la machine virtuelle

Vous pouvez accéder à la VM via la console VNC configurée, ou utiliser SSH si vous avez configuré l’accès SSH dans cloud-init (vérifiez l'IP dans virsh ou avec virsh domifaddr debian12).
Récapitulatif

    Préparer les fichiers cloud-init (user-data et meta-data).
    Créer l’ISO cloud-init avec genisoimage.
    Créer le fichier XML de description de la VM.
    Définir et démarrer la VM avec virsh.

Cela devrait être suffisant pour instancier une machine virtuelle Debian 12 initialisée avec cloud-init sur votre système Debian utilisant libvirt et qemu-kvm.