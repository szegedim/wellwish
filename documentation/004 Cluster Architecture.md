# Cluster Architecture
### Author: Creative Commons Zero
### Date: 2023-05-26

The design is the result of three years of extensive research from 2020-2023 by Schmied Enterprises.

Productivity is the measure of the domestic product achieved per employee.
It has not improved much in decades, once data centers became mainstream.
The goal was to address the problem of productivity, with a simple, open benchmark solution.

The main concept behind WellWish is that has extremely low support costs.
The marginal labor cost of running a cluster per node does not increase by the cluster size.
It requires no-dev, no-ops, no-os. The concept is also known as the personal cloud.

There are two ways to achieve this.
One is that it uses Englang, plain words to describe code and data.
The second is that the entire state can easily be retrieved, stored, analyzed and ported in Englang.
Even an accountant can read the bare metal data files that are equivalent to logs.

Kubernetes can do the same with pods that implement it.
The main difference is the mesh structure.
Kubernetes has a master, while Wellwish distributes even administrator requests
across a unified cluster called the office.
Each node in an office contains all modules in a proportion as many resources they require.
The service hey.com, the email service has similar architecture.

Nodes of the same type scale better and cheaper.
Resource load differences between microservices are set within each node independently vs. the entire cluster.
Therefore, each node has a stateful container and many stateless burst containers that pick up requests and restart.
This structure helps to choose the most efficient node types of major cloud providers.
It also does not require engineering to decide on scaling policies.

There are no roles. Each room in the office is reserved for data container or burst code.
Each of them has a unique private apikey that you can use to knock and use that room.
This concept is similar to the original Disk Operating System sold by Microsoft decades ago
in the age of your great-grandfather.

The main business advantage is low support cost.
Each structure is allowed to be twice as large as a classical binary or json.
This is not a big deal in the 2020s.
This allows us to use Englang - Engineering English Language.
Englang writes code in words. If in doubt, the rules of the English language apply. 
This help users, accountants, debuggers, etc.
Doubling the buffer size still scales well.

While it is designed to be scalable, we suggest using a cluster size of two nodes for optimal reliability.
A single node does not ensure scaling when it is needed.
One thousand nodes may bring in node errors with lost or delayed shards,
where some customers end up on older versions.
Two nodes ensure that any node errors surface quickly, and it triggers a drill to fix fast.
Two nodes ensure that scaling works, and it is easy to add a third node when needed.

Network issues can surface as a bigger fixed cost with larger clusters,
making support staff busy with low impact sporadic errors. 

This cluster architecture is stateful versus Kubernetes.
Data is stored mostly in memory, and it backs up fast to real stateful backup storage.
This allows quick offloading and shutdown of extra nodes, when room use is sporadic like at night.
See Storage and Database for details.