# TCP/IP Stack Simulation

This project is a simple implementation of a TCP/IP stack simulation, focusing on Layer 2 and Layer 3 functionalities. It allows you to build network topologies, set routing entries, and perform basic networking operations.

## Getting Started

### Build the Project

To build the project, use the following command:

```bash
make build
```

### Run the Project

Run the compiled executable with the following command:

```bash
./tcpip
```

## Setting Routes

You can set routing entries using the `config node route` command. Here is an example:

```bash
config node route R1 122.1.1.3 32 10.1.1.2 eth0/1
```

## Ping Operation

To ping a destination address, use the `run node ping` command. Example:

```bash
run node ping R1 122.1.1.3
```

You can also perform IP-in-IP encapsulation using the following command:

```bash
run node ping tunnel R1 <destinationIP> <tunnelIP>
```

## Additional Functionalities

- **VLAN Support:** Create and manage VLANs for network segmentation.
- **Loopback Address:** Assign loopback addresses to nodes for local testing.

## Topology Customization

In the `topology/topology.go` file, you can create your own topology. Use the created topology in the `cmd/commands.go` file by calling the function you created to replace the variable `topology` with the returned topology.

## Future Development

This simulation is a work in progress, and more features will be added in the future.

Feel free to explore the code and experiment with different network topologies!
