# Distributed KVM Switch

## About the project
The Distributed KVM Switch (dKVMs) is a framework that aims to manage input and output streams from keyboards, mice, displays, speakers, and microphones in a distributed manner. The final objective of the project is to create a platform-independent distributed KVM switch, similar to the hardware one.

## Current state
The project is currently in its initial stages, with ongoing development and potential features yet to be implemented, such as supported devices. Check the "issues" page to track progress and to report any issues.

### Supported platforms
The goal of this project is to provide support for three major platforms, namely MacOS, Windows, and Linux. To use this project, an agent must be installed, which requires root/administrator privileges.

However, in a corporate environment, it may be challenging to use the project since root/administrator permissions are not always available for a user account. To overcome this limitation, we are planning to introduce an agentless mode that will operate through Raspberry Pi as an agent. You can find more details about this in the section below.
#### MacOS
> TBD

#### Windows
> TBD

#### Linux
> TBD

#### Raspberry Pi
> TBD

## Project architecture

### Overview
The project involves two important concepts that work together to ensure proper functioning - the kernel and the devices. The kernel is responsible for managing various devices and processing streams of data that originate from these devices. Essentially, the kernel starts by reading the configuration file and then proceeds to start the devices. Each device is an operating system process and is controlled by the kernel through a wire protocol over stdin/stdout pipes. The devices produce events that contain data to be processed, such as video frames, keystrokes or mouse movements. The kernel is responsible for routing the events to other devices as required.

While the kernel and devices may seem complex, their functionality is relatively straightforward. If you require more information, please refer to the section below for additional details.

### Kernel
The kernel has several responsibilities, and each of then is handled by separate module.

#### `device-manager`
The kernel has a major responsibility of managing devices, which is done through the device manager module. This module is responsible for setting the desired state of each device, as defined in the configuration file after startup. Since a device is just an OS process, the device manager module also starts and stops the device processes.

In addition, the device manager module monitors the current state of the device process and restarts it if it is not running. When a device process completes with an exit code other than 0, it is restarted with exponential backoff. This module also attaches wire protocol to the device pipes (stdin and stdout) and connects handlers from other modules to the wire protocol.

Furthermore, the device manager module connects a log handler to the stderr pipe, which captures all log messages and processes them in the same way as logs from the kernel. 

Like several other modules in the kernel, the device manager module exposes services that can be utilized by the device. This service provides APIs for querying the state of the devices. It also emits events when the device state changes, and devices can subscribe to these events.

#### `memory-manager`
Modern devices often produce large amounts of data, which can be quite challenging to handle. For instance, to capture a raw frame of Full HD display, the device responsible for streaming would need to emit 60 buffers with frames, producing about 355 MB/s of data. This data would then need to be passed to a device that displays it on the screen, or compressed and passed over a network to another kernel, where it can be decompressed and then displayed on the screen. This can result in a lot of data shuffling.

One might think that sharing pointers instead of copying data could solve this problem. However, due to the nature of the device ([separate process](#device)) and modern operating systems ([virtual memory](https://en.wikipedia.org/wiki/Virtual_memory)), it is not possible to share a pointer to memory from one process to another. Every process has its own linear memory space, which cannot be easily shared with another process.

Fortunately, operating systems allow us to request shared memory space, which is typically exposed as a file descriptor that can be easily mapped inside the process as a place in virtual memory space. This enables us to pass data between different processes. However, this shared memory is just a simple, plain large buffer; therefore, we need to bring some management to it. This is where the memory manager comes in handy.

The memory manager is responsible for allocating buffers in shared memory space, locking them, retaining, releasing, and so on. It's like a regular memory manager in an operating system, but only for shared memory space. Like other modules in the kernel, it exposes services that can be utilized by the device. Therefore, when a device needs to pass data to another device, it should first request the memory manager to allocate a buffer in shared memory space, write data to that buffer, and then pass only the buffer descriptor to the outgoing event.

### Device
As you may have read before, the device is essentially an OS process. You might be wondering why a device is not just simple code that can be used by the kernel? That approach would be much simpler rather than creating separate processes, managing them, their configurations, wire protocol, and so on. Why is this separation level used? It seems to complicate things. It is worth understanding why devices are separate processes.

Devices typically use different internal and public OS APIs. Sometimes, these APIs work unpredictably and cause the calling process to hang or crash. Unfortunately, exception handlers can't catch those crashes because these errors are often segmentation faults or other low-level violations that cause the process to be killed. To avoid crashing the whole project, it is better to move the device logic to a separate process. If that separate process crashes, the device manager will restart that process, and the device will be initialized again.

Another reason for this separation is that calling OS kernel APIs (e.g. to install a keyboard hook or access screen capturing service) usually requires being done from the main thread of the caller process. Combining main thread access to MacOS APIs and MS APIs, in the same way, would be super hard. It is better to allow the use of native GCD or another native event loop mechanism instead of combining different OS APIs.

#### SDK
> TBD

### Supported devices
> TBD