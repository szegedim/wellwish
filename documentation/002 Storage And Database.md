# Storage And Database
### Author: Creative Commons Zero
### Date: 2023-05-19

## Basic storage

There is no good storage solution in case of cloud native solutions.
Kubernetes was started as a computing cluster.
Any volume and storage options were added later, but they are way too complex.

Our solution is an in memory streaming and caching finished on disk stateful backup, if needed.

## Responsiveness

The main design concept is responsiveness.

Assume that a cluster has hundred million users.
This means that the per-user active transactional data is like twenty megabytes.
Some users may have more but the customer base suggests a small amount.

This can easily be sharded by user buckets to individual servers in string to string apikey memory maps.

We use the apikey as a sufficient solution for indexing.
Once the server is chosen, serving the data is quick from memory.

Indexing is simple by exchanging the keys only in a ring.
The index is a map from the apikey to the server that contains the relevant data.
More complex data structures can be established at the application level mixing the keys.

## Long term storage

Stale buckets and stale users can then be cleaned up periodically.
This makes our computing cluster cache only.

The backup solution has a push and pull behavior.
We can have more of these backup streaming pipelines filtered by the module that is backed up.

Some modules may be more critical, where the application model itself requires and implements replication.
Those modules can just replicate a slot with new keys that likely end up on another nodes.

Many modules do not need replication. Data lost can be just reentered every two years in case of a loss.
Temporary modules like chat indexes can safely be deleted.
Columnar indexes can safely be deleted, because recreating them is cheap and straightforward from primary data.

Indexes are backed up to be addressed, but they are never offloaded from memory.

Periodic stateful backups can offload some data from servers.
They can also be served as restore solution, if the user data is accessed again.
If the index is not found, the systems asks the backup.

The stateful backup server acts together with the compute cache.
The final source of truth is the stateful backup server, computing is cleaned regularly.

## Privacy

This is a very affordable Privacy and Law driven design.
The design makes it very easy and straightforward to fulfil privacy and compliance requirements like GDPR.

Non-essential private data is cleaned up periodically in an hour or in a week.
Permanent data has one single location on a stateful backup to locate easily.
In-memory caches are recreated periodically, so any deletion request is fulfilled by design deleting the backup.

Any extensive streaming data collection cannot be done with this concept but that may not be needed in the 2020s.
Our solution on the other hand is simple, very fast, and compliant for production, money generating workloads.

## Workflow

We use the same facility to back up, restore, stream, and to replicate.

We have a separate storage system only for stateful backups.

Users enter data directly to stateless in memory containers bucketed by api keys.
These are temporary state containers. In memory state is also a state, it is just transient.

Some data is simply deleted after two minutes, a day or a week, when it is not necessary anymore.
Some data may be offloaded or backed up by a periodic request to the stateful backup system.

Failed pods or containers can get requests that they cannot fulfill.
They can just fetch the missing data to memory from a stateful backup server.
The backup server access information is periodically pushed and updated at the time of snapshots.

The stateful backup system monitors and downloads new data.
This means that it fetches everything.
It is not such a big deal, since memory cloud containers are expensive compared to storage containers.
Stateless container content will be a fraction of the stateful container size.
We fetch only what is frequently used as a result.
Offloading everything makes the logic simple, and it will allow consistency checks on the indexes catching malware.

Any critical workloads can be made safer by making the stateful backup period shorter.
Such short stateful backups can be incremental snapshots getting just what has changed.

## Cloud provider backups

Stateful backup servers have larger disks. They are typically not microservices but entire VMs or bare metal servers.
They can be backed up simply by cloud provider specified backup solutions.
Physical backups can be done this way, but only when is needed.
Physical backups can be the most likely sources of data leaks, as they are watched rarely.

## Backup format

The author once attended a conference selling enterprise database solutions.
An activist investor came asking about the portability of the data.
Investors care about competition. It makes solutions affordable for their investment.
It increases the pace of growth and profitability for portable systems vendors.

The backup format is plain Englang (a.k.a English) key value pairs by apikey.
This makes the raw data files readable even by accountants and lawyers.

Nothing is an unreadable binary blob.
It is also very elastic and scales up and down easily.
Buckets can be loaded back to other pods than where they originated from.
Offloading and load balancing comes for free.

## Debugging

The system can be debugged by checking the debug logs that come together with the backup.
This means that any transactions are equivalent to the logs that created them.
You can restore entire transactions by replaying the logs.

Any indexes are based on the backed up logs.
They are recreated periodically.

There is no noise or inconsistency by design.
ACID is part of the rebuilding of temporary indexes.
This is an Available system by design based on the CAP theorem.

Modules can implement application specific replication.
This is very streamlined and affordable this way.

Modules implement their own consistency rules, obviously.
The api key based zero trust solution helps a lot to achieve isolation.

The logs are plain Englang - English.
Only some algorithms such as bitmap indexes need binary solutions that can be just temporary caches anyway.

## The Science

**Problem**
Such a mixed stateless & stateful solution is optimal in terms of latency, the resources given in advance.

**Solution**
The solution uses an in memory cache as the primary storage, and streams.
The data uses the fastest storage early. It is offloaded only when memory is full.
Any other solution involving non-volatile memory would be slower than the dynamic memory.
A sufficient offload selection algorithm like Least Recently Used can make sure that the algorithm is latency optimized.

Systems that mix stateful and stateless storage on the same nodes like Hadoop are rare.
They are useful in a narrow range of workloads accessing tremendous amounts of data with interactive queries.
This is rare since companies naturally optimize for cost keeping only valuable data.
Our stateless memory interfaces are more responsive as a result.
We also do not add the burden of fixed cost configuration of Kubernetes YAMLs, Helm charts, Secrets, Roles, VPNs, Ingresses, etc. 

Up scaling can still be achieved easily.
Let's assume a photo store for those hundred million customers.
The system would offload large files to those stateful servers immediately keeping only what is needed for a quick startup.
Image processing can be streamed to stateless containers.
Large image processing can be a separate computing cluster from and to the stateful backup machines for free.
They are always up-to-date after a definite backup period.
