# Rapide Advanced Parallel protocol Indiferent Download orchEstrator

The goal is to write a fast heavly parallel (capable to use the bandwidth of many nodes)
bitswap and graphsync composite client.

It work by doing pseudo-workstealing of blocks downloads with a DFS traversal.

# How it works

## Bitswap

Bitswap runs as an event loop.

It has datastructure to locate `peer.ID, blocks.Block` back into positions in the DAG.

When a message is received the DAG position is recovered and then it march the traversal.

This code would be a great candidate for a C or Zig translation. Awful for a Rust one.

# Licensing

This project is licensed under the MIT license, see [LICENSE](LICENSE).
