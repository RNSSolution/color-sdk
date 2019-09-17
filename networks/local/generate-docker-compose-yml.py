#!/usr/bin/python3

import ipaddress
import argparse
import sys

parser = argparse.ArgumentParser(description='Generate docker-compose.yml')
parser.add_argument("leagues", type=int, help="number of leagues")
parser.add_argument("nodes", type=int, help='number of nodes in each league')
parser.add_argument("network", type=ipaddress.IPv4Network, help="local network that the nodes belong to.")
parser.add_argument("ipaddr", type=ipaddress.IPv4Address, help="starting IP address of the nodes array")
parser.add_argument("port", type=int, help="starting port for forwardint nodes' RPC ports")
parser.add_argument("-t", "--tendermint", action='store_true', help="starting port for forwardint nodes' RPC ports")

def print_version():
	print("version: '3'")

PRISM_NODE_TEMPLATE="""  node{id}:
    container_name: colornode{id}
    image: "tendermint/colordnode"
    ports:
      - "{startport}-{endport}:26656-26657"
      - "{profport}:6060"
    environment:
      - ID={id}
      - LOG=colord.log
      - COMMAND_ARGS
      - SUBCOMMAND_ARGS
    volumes:
      - ./build:/colord:Z
    networks:
      localnet:
        ipv4_address: {ipaddr}"""

TM_NODE_TEMPLATE="""  node{id}:
    container_name: node{id}
    image: "tendermint/localnode"
    ports:
      - "{startport}-{endport}:26656-26657"
    environment:
      - ID={id}
      - LOG=tendermint.log
    volumes:
      - ./build:/tendermint:Z
    networks:
      localnet:
        ipv4_address: {ipaddr}"""

NODE_TEMPLATE = None

def print_node(i, ipaddr, port, profport):
	nodespec = NODE_TEMPLATE.format(id=i, startport=port, endport=port+1, ipaddr=ipaddr, profport=profport)
	print(nodespec)
	print()

def print_services(leagues, nodes, ipaddr, port, profport):
	print()
	print("services:")
	for i in range(leagues*nodes):
		print_node(i, ipaddr+i, port+2*i, profport+i)

NETWORKS_TEMPLATE = """networks:
  localnet:
    driver: bridge
    ipam:
      driver: default
      config:
      -
        subnet: {network}"""

def print_network(network):
	spec = NETWORKS_TEMPLATE.format(network=network)
	print(spec)
	print()

def run():
	args = parser.parse_args()
	fail = False

	if args.leagues <= 0:
		print("Number of leagues must be positive", file=sys.stderr)
		fail = True
	if args.nodes <= 0:
		print("Number of nodes must be positive", file=sys.stderr)
		fail = True

	if not args.ipaddr in args.network:
		print("Starting IP address doesn't belong to the network", file=sys.stderr)
		fail = True
	elif not args.ipaddr+args.leagues*args.nodes in args.network:
		print("Ending IP address doesn't belong to the network", file=sys.stderr)
		fail = True

	if args.port <= 0:
		print("Port must be positive", file=sys.stderr)
		fail = True
	elif args.port+2*args.leagues*args.nodes > 65535:
		print("Port is too big: all ports must fit in [0,65535] range.")
		fail = True

	global NODE_TEMPLATE
	if args.tendermint:
		NODE_TEMPLATE = TM_NODE_TEMPLATE
	else:
		NODE_TEMPLATE = PRISM_NODE_TEMPLATE

	if fail:
		sys.exit(1)

	print_version()
	print_services(args.leagues, args.nodes, args.ipaddr, args.port, 60600)
	print_network(args.network)

if __name__ == '__main__':
	run()