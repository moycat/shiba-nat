# Shiba NAT

This is a NAT tool for [Shiba](https://github.com/moycat/shiba), enabling IPv6-only nodes to access IPv4 network.

The IPv4 traffic is routed to dual-stack nodes and get NAT-ed.

## Installation

1. Label gateway nodes with `shiba/nat=gateway` and IPv6-only nodes with `shiba/nat=client`.
2. Run `kubectl apply -f installation.yaml` and enjoy.
