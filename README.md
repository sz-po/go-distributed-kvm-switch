# Distributed KVM Switch

## About the project
The Distributed KVM Switch (dKVMs) is a framework aimed at managing input and output streams from keyboards, mice, displays, speakers, and microphones in a distributed manner.

## Current state
The project is currently in its initial stages, with ongoing development and potential features yet to be implemented, such as supported devices. Check the "issues" page to track progress and to report any issues.

## Project architecture

### Overview
The project brings two concepts: _the kernel_ and _the devices_. To put it very simply, the kernel manages devices and processes streams from the devices. When the kernel starts, it reads the configuration file and starts the devices. The device is just an OS process, controlled by the kernel via wire protocol over stdin/stdout pipes. The device produces events that contain data to be processed (e.g. video frame, keystroke, or mouse movement), and the kernel is responsible for routing events to other devices.

Kernel and devices are more complex than presented here, but not too much. Read the section below for more details.

### Kernel
The kernel has several responsibilities, and each of then is handled by separate module.

#### `device-manager`
One of the main responsibilities of the kernel is to manage devices. This module allows setting the desired state of each device (which is done via the configuration file right after startup). As the overview section said, the device is just an OS process, so this module starts and stops the device's processes.

It is also responsible for watching the current state of the device process and restarting it if it is not running. For example, when the device process finishes with an exit code other than 0, it will be restarted with exponential backoff. This module is also responsible for attaching wire protocol to the device pipes (stdin and stdout) and attaching handlers from other modules to the wire protocol.

Moreover, it attaches a log handler to the stderr pipe, to catch all log messages and process them in the same way logs from the kernel.

This module, like several other ones in the kernel, exposes services that can be used by the device. This service provides APIs for querying of state of the devices. It also emits events when the device state changes and devices can subscribe to these events.

### Device
As you likely read before, the device is just an OS process. Maybe you're thinking:
> Why device is not simple code that can be used by the kernel? That approach will be far simpler instead of spawning separate processes, managing them, their configurations, wire protocol, and so on. Why that separation level is used? It complicates everything. WTF.

At first glance, it may look strange, because this architectural decision brings some complications mentioned before. But first, let's explain why devices are separate processes.

Typically, the devices utilize different internal and public OS APIs. Sometimes those APIs work unpredictably and cause the calling process to hang or crash. Unfortunately, exception handlers can't be used for catching those crashes, because very often those errors are segmentation faults or other low-level violations that cause process kill. To avoid crashing the whole project is better to move the device logic to a separate process. If that process crashes, the device manager will restart that process, and the device is initialized again.

Also, there is another reason. Calling OS kernel APIs (e.g. to install a keyboard hook, or access screen capturing service), usually requires to be done from the main thread of the caller process. It will be super hard to combine main thread access to MacOS APIs and MS APIs in the same way. It's better to allow the use of native GCD or another native event loop mechanism vs combining different OS APIs. 

> TBD

### Supported devices

> TBD