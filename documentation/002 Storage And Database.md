# Storage And Database
### Author: CC0
### Date: 2023-05-19

## Basic storage

There is no good storage solution in case of cloud native solutions.
Kubernetes was started as a computing cluster.
Any volume and storage options were added later, but they are way too complex.

Our solution is an in memory streaming solution finished in on disk stateful backup, if needed.

## Responsiveness

The main design concept is responsiveness.

Assume that a cluster has hundred million users.
This means that the per-user active transactional data is like twenty megabytes.
Some users may have more but the customer base suggests a small amount.

This can easily sharded by user buckets to individual servers in string to string apikey memory maps.

We use the apikey as a sufficient solution for indexing.
Once the server is chosen, serving the data is quick from memory.

Indexing is simple by exchanging the apikeys only in a ring.
The index is a map from the apikey to the server that contains the relevant data.

## Long term storage

Stale buckets and stale users can then be cleaned up periodically.
This makes our computing cluster is a cache only.

The backup solution has a push and pull model.
We can have more of these filtered by the module that is backed up.
Some modules may be more critical, where the application model can increase replication.
Many modules do not need replication. Data lost is just reentered every two years if this happens for example.

Periodic stateful backups can offload some data from servers.
They can also be served as restore solution, if the user data is accessed again.

The stateful backup server together with the compute cache act together.
The final source of truth is the stateful backup server, computing is cleaned regularly.

## Privacy

This makes it very easy and straightforward to fulfil GDPR like privacy compliance requirements.
This is a very affordable Privacy and Law driven design.

Any extensive streaming data collection cannot be done with this concept but that may not be needed in the 2020s.
Our solution on the other hand is simple, very fast, and compliant for production, money generating workloads.

## Workflow

We use the same facility to back up, restore, stream, and replicate.

We have a separate storage system only for stateful backups.

Users enter data directly to stateless in memory containers bucketed by api keys.
These are honestly temporary state containers. In memory state is also a state, it is just transient.

Some data is simply deleted after two minutes, a day or a week, when it is not necessary anymore.
Some data may be offloaded or backed up by a request from the stateful backup system.

Failed pods or containers can get requests that they cannot fulfill.
They can just fetch the missing data to memory from a stateful backup server.
The backup server access information is periodically pushed and updated at the time of snapshots.

The stateful backup system monitors and downloads new data.
This means that it fetches everything.
Any critical workloads can be made stronger by making the stateful backup period shorter.
Such short stateful backups can be incremental snapshots getting just what has changed.

## Cloud provider backups

Stateful backup servers have larger disks. They are typically not microservices but entire VMs or bare metal servers.
They can be backed up simply by cloud provider specified backup solutions.
Physical backups can be done this way, but only when is needed.
Physical backups can be the simplest sources of data leaks.

## Backup format

The backup format is plain Englang - English key value pairs by apikey.
This makes it readable even by accountants.
It is also very flexible. Buckets can be loaded back to other pods where they originated from.
Offloading and load balancing comes for free.

## Debugging

The system can be debugged by checking the debug logs that come together with the backup.
This means that any transactions are equivalent to the logs that created them.
You can restore entire transactions by replaying the logs.

There is no noise or inconsistency by design.
ACID is part of the rebuilding of temporary indexes.
This is an Available system by design based on CAP theorem.

Modules can implement application specific replication.
This is very streamlined and affordable this way.

Modules implement their own consistency rules, obviously.
The api key based zero trust solution helps a lot to achieve this.

The logs are plain Englang - English.
Only some algorithms like bitmap indexes need binary solutions that can be just temporary caches anyways.
