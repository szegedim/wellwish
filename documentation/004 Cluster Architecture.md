# Cluster Architecture
### Author: Creative Commons Zero
### Date: 2023-05-26

## Service Mesh

We have a different approach to service mesh than the traditional one.

A load balancer is a cloud or datacenter equipment that randomly spreads requests across
a unified set of backend nodes.
Sometimes it can be set sticky to refer to the same server that fulfilled a request already.
This helps performance in general.
Unfortunately stickiness usually is by IP or cookies. We advise against these.

Our office is behind a load balancer and firewall in a clustered scenario.
The data is stored only on a single server.
This makes sense.
Application layer logic can easily replicate important data sets due to the simple room design.

When a node is hit, it uses an index to find the node that has the data.
This index is repeatedly kept up to date with new items.
The surface node just proxies the request instead of forwarding with a redirect.
The reason is that redirecting may cause issues in a load balancer.
Node addresses are not easily discoverable outside the cluster.

## Fully stateless modules with load balancing

Modules can be fully stateless, in which case they just respond without any load balancing issues.

## Interactive real-time modules with ring indexing

Interactive workloads with strong response time requirements need stickiness with indexing.
Each node forwards its local indexes to the next server in a ring architecture.

Indexing discovers and orders all the nodes by a predefined pattern.
This is set by the NodePattern setting in the metadata.
Then each time a node in the ring wakes up it pings the servers on the right to find the first active one.
We allow elasticity and cluster discovery very easily this way.
Gaps in the node list are just ignored.
The Nodepattern can be ip address with a cidr like 10.55.0.0/21 or a host name pattern like host**.example.com.
The wildcard is then replaced with a generated pattern of digits 0-9 to generate 100 node addresses.
Only nodes that return success on a /health request do participate.

## Stateful big data workloads with snapshot and backup backend

Some workloads like storing images or videos may need regular backups.
The storage subsystem can be activated for these modules.
This one offloads and deletes some least recently used items.
Sometimes the data is offloaded to a backup already.
Modules can fetch the data from the stateful servers in this case.
Refer to StatefulBackupUrl setting for details.

## Performance

A typical container has a bandwidth comparable to 1 vCPU/1GB RAM/1Gbps.
The benchmark is that it can be offloaded and shut down easily.
The network allows this for 1GB RAM in about ten seconds.
Large data blobs can be offloaded to backups with snapshots of the same bandwidth characteristics.

This means that a ten seconds indexing frequency is suitable for most applications.
A shared document review of two team members sitting next to each other is a good example.

The module should have a single entry point apikey.
Any browser refresh would give the same document for the same key on the same node by the usual default IP stickiness.

If the browser url is shared with the peer, it is normal that they wait for five seconds after opening some new windows to get to the document.
A ten seconds indexing frequency fulfils this requirement, even if the ip address of the recipient is different, than the ip of the sender.

## Future designs to consider

Smart load balancers can be better by being sticky by the apikey.
Such a next generational load balancer can eliminate this indexing logic.

Large indexes can be handled by storing only the first few characters of the api keys in the hash maps.

Data retention or indexing usage flags can be set by each module.

An important fact is that this design allows easy compliance with privacy regulations.
The module is application level knowing the scope and sense of the data and whether it is personal or not.

## License

```
This document is Licensed under Creative Commons CC0.
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
to this document to the public domain worldwide.
This document is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this document.
If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.
```
