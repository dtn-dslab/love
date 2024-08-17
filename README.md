# LoVe: Low overhead VEth APIs

## What is LoVe?

LoVe is a set of low overhead VEth APIs that can be used to create and manage VEth interfaces. LoVe considers the overhead of relating _netlink_ messages to the VEth interfaces and provides a simple and efficient API to manage VEth interfaces.

## Why LoVe?

LoVe is designed to perform well under high load and to be simple to use. It is designed to be used in high-performance applications(such as Network Emulation) that require the creation and management of VEth interfaces.

Under specific conditions, the overhead of creating and managing VEth interfaces using the _netlink_ API can be significant. Compared to [koko](https://github.com/redhat-nfvpe/koko), LoVe performs 10x better in terms of the time taken to create and manage VEth interfaces.

## Design

Despite some operations bring much overhead, LoVe is still capable of providing as many functions as possible.

LoVe divides functions to remind users of the overhead of each operation. LoVe also wraps the structs provided by the _netlink_ package to provide a more powerful and easy-to-use API.

## Features

