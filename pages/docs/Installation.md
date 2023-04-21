---
sidebar_label: 'Installation'
sidebar_position: 2
---

# Installation

## Requirements

:::caution
Install first :
* Qemu-KVM & Libvirt
* Python >= 3.9
:::

## Install

Clone Freyja from the Git sources :

```sh
git clone https://gitlab.thalesdigital.io/theresis/freyja && cd freyja
```

Run :

```sh
pip install dist/freyja-0.1.0-py3-none-any.whl
```

:::info Upgrade
In the same way, upgrade freyja with `pip install --upgrade dist/freyja-0.1.0-py3-none-any.whl`
:::