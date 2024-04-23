# Distributed KVM Switch

## About the Project
The Distributed KVM Switch (dKVMs) is a framework designed to process input and output streams from keyboards, mice, displays, speakers, and microphones in a distributed manner.

## Current State
The project is in its early development stage. This section outlines the tasks to be completed and tracks the project's progress.

### Kernel
- Memory
    - Implement shared memory buffer between the kernel and devices.
    - Develop release/retain mechanism.
    - Enable zero-copy buffer passing between the kernel and devices.
- Device Management
    - Implement starting and stopping devices, with retry on failure and backoff mechanism.
    - Enable passing configuration to devices based on a YAML file read on kernel start.
    - Implement collecting logs from devices stderr stream.
- Pipelines
    - Load pipeline description from a YAML file.
    - Support simple pipelines to facilitate passing events from one device to another.

### SDK
- Wire Protocol (based on stdin/stdout pipes)
    - Develop event sending and receiving mechanisms.
    - Enable Remote Procedure Call (RPC) functionality for calling the kernel and devices.
- Configuration
    - Allow passing configuration through environment variables.
- Logging
    - Implement logging mechanism to ship logs to the kernel using stderr.
- Data Handling
    - Enable requesting data buffers from the kernel and passing them over the wire.

### Devices
Here is a list of devices that should be implemented to enable other developers to create their own devices.

#### `dummy-display-source`
- Generate random output frames.
- Generate predefined frames previously read from a file.
- Support partial buffer.

#### `dummy-display-sink`
- Save frames to files.
- Support partial buffer.

#### `dummy-keyboard-source`
- Generate predefined keyboard events from a file.

#### `dummy-keyboard-sink`
- Save keyboard events to files.

#### `dummy-mouse-source`
- Generate predefined mouse events from a file.

#### `dummy-mouse-sink`
- Save mouse events to files.

#### `dummy-audio-source`
- Generate sine waves.
- Generate predefined audio waves from a file.

#### `dummy-audio-sink`
- Save audio waves to files.

#### `tcp-server`
- Allow connection of other kernels and passing events.

#### `tcp-client`
- Allow connection to other kernels and passing events.

## Design Concepts
The dKVMs consists of two main components: the kernel and devices. The kernel serves as the core process responsible for spawning and managing devices, as well as executing pipelines. Devices, on the other hand, act as producers or consumers of input and output streams. Additionally, devices are capable of sending events to the kernel and reacting to them.