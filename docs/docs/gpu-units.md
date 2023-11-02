---
layout: default
title: GPU processing inside Bravetools
parent: Docs
nav_order: 6
description: "Enable GPU-accelerated processing inside Bravetools Units"
---

# GPU Processing Inside Bravetools
Many machine learning applications require GPU-accelerated hardware. Although many personal computers and servers ship with GPU support, it can be desirable to scale training operations to multi-GPU systems in cloud deployments. Bravetools uses [LXD](https://documentation.ubuntu.com/lxd/en/latest/) to configure and manage GPU hardware inside your Units. A key advantage of this approach is that you can scale your training scripts and environments from a local machine to massive distributed systems without changing your configuration code.

In this section, we will describe how to enable NVIDIA hardware inside your Units.

## Before you begin
It's likely that some housekeeping work will need to be carried out, such as ensuring that old NVIDIA drivers and __nouveau__ drivers are removed. For details refer to [this excellent resource](https://ubuntu.com/tutorials/gpu-data-processing-inside-lxd#2-remove-nvidia-drivers).

## System requirements

* NVIDIA Drivers
* [CUDA toolkit](https://developer.nvidia.com/cuda-downloads)

To ensure your system components are set up correctly, run `nvcc -V`. If CUDA is installed correctly, you should see output like the following with your CUDA installation information:

```bash
nvcc: NVIDIA (R) Cuda compiler driver
Copyright (c) 2005-2020 NVIDIA Corporation
Built on Thu_Jun_11_22:26:38_PDT_2020
Cuda compilation tools, release 11.0, V11.0.194
Build cuda_11.0_bu.TC445_37.28540450_0
```

Finally, if CUDA was installed with extras included, you may ensure all NVIDIA components are functioning correctly by running `$CUDA_HOME/extras/demo_suite/bandwidthTest`. Typically, CUDA is installed in `/usr/local/cuda` on Linux systems. The output should contain `Result = PASS`.

## Enable GPU support in a Bravefile

GPU support can be easily enabled inside the **Service.Resources** section of a [Bravefile](../bravefile) by setting `gpu: "yes"`:

```yaml
service:
  image: ubuntu-bionic-gpu/1.0
  name: ubuntu-bionic-gpu
  docker: "no"
  version: "1.0"
  ip: ""
  ports: []
  resources:
    ram: 4GB
    cpu: 2
    gpu: "yes"
```

After running `brave deploy`, your newly created Unit will be able to utilise your system's NVIDIA configuration to tap into GPU acceleration!