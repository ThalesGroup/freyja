---
title: 'Python requirements'
sidebar_label: 'Python & Poetry'
---

# Install Python 3 and Poetry

## On Ubuntu

Check python 3.9 :

```sh
python3 --version
```

If python 3.9 is not installed :

```sh
sudo apt install software-properties-common
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt install python3.9
```

Install Poetry :

```sh
curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python3.9 -
source $HOME/.poetry/env
poetry --version
```
