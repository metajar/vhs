# VHS - Network Device Backup System

VHS (Virtual Host System) is a backup system for network devices, designed to automatically save configurations and related data to a remote Git repository. This system is an essential tool for network administrators, helping them to maintain a history of device configurations and simplifying the process of restoring previous settings in case of failures or misconfigurations.

## Features

- Automatically saves configurations of network devices to a Git repository
- Supports periodic pushes to the remote repository
- Deprecates old configuration files after a specified time period
- Redacts sensitive data from the saved configurations
- Provides an example client to interact with network devices over SSH
- Implements a simple and efficient server using the Twirp framework

## Prerequisites

- Golang 1.16 or higher
- Git command-line tool installed and configured

## Installation

1. Clone the VHS repository:

```sh
git clone https://github.com/metajar/vhs.git
```

2. Change the working directory to the `vhs` folder:

```sh
cd vhs
```

3. Build the server binary:

```sh
go build -o vhs-server ./server
```

4. Build the example client binary:

```sh
go build -o vhs-client ./client
```

## Usage

### Server

1. Update the server configuration in the `server.go` file with your Git repository URL and branch name.

2. Run the server:

```sh
./vhs-server
```

The server will be accessible at `http://localhost:8080`.

### Example Client

1. Update the client configuration in the `client.go` file with the IP address, username, and password for the network device you want to backup.

2. Run the example client:

```sh
./vhs-client
```

The client will connect to the network device, execute a set of commands to retrieve the device's configuration, and send the configuration to the VHS server.

## Customization

You can customize VHS by modifying the server or client code to support different types of network devices or additional features. You can also create your own client applications using the provided `VhsServiceProtobufClient` API.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any bugs, improvements, or feature requests.

## License

VHS is licensed under the MIT License.