# WellWish Corporate Decision Engine

Wellwish is a platform similar to Kubernetes.

It allows you to run applets within your company that support decision-making.

## Design

The design is the result of three years of extensive research from 2020-2023 by Schmied Enterprises.

Productivity is the measure of the domestic product achieved per employee.
It has not improved much in decades, once data centers became mainstream.
The goal was to address the problem of productivity, with a simple, open benchmark solution.

The main concept behind WellWish is that is extremely low support costs.
The marginal labor cost of running a cluster per node does not increase by the cluster size.
It requires no-dev, no-ops, no-os. The concept is also known as the Personal Cloud.

There are two ways to achieve this.
One is that it uses Englang, plain words to describe code and data.
The second is that the entire state can easily be retrieved, stored, analyzed and ported.
Even an accountant can read the bare metal data files.

Kubernetes can do the same with pods that implement it.
The main difference is the mesh structure.
Kubernetes has a master, while Wellwish distributes even administrator requests across a unified cluster called the office.

Nodes of the same type scale better and cheaper.
Resource load differences between microservices are set within each node independently vs. the entire cluster.
Therefore, each node has a stateful container and many stateless burst containers that pick up requests and restart.
This structure helps to choose the most efficient node types of major cloud providers.

There are no roles. Each room in the office is reserved for data or burst code.
Each of them has a unique private apikey that you can use to knock and use that room.

The main business advantage is low support costs.
Each structure is allowed to be twice as large as a classical binary or json.
This allows us to use Englang - Engineering English Language.
This help users, accountants, debuggers, etc.
Doubling the buffer size still scales well.

While it is designed to be scalable, we suggest using a cluster size of two nodes for optimal reliability.
A single node does not ensure scaling when it is needed.
One thousand nodes may bring in node errors with lost shards, where some customers end up on older versions.
Two nodes ensure that any node errors surface quickly, and it triggers a drill getting fixed fast.
Two nodes ensure that scaling works, and it is easy to add a third node when needed.
Network issues can surface as a bigger constant cost with larger clusters,
making support staff busy with low impact sporadic errors. 

## Who is it for?

Wellwish targets a specific user base.

Creative Commons open source is suitable for biotech, healthcare, robotics research and development businesses, who are patent holders themselves.
Some open source copyright licenses other than Creative Commons may pose a patent risk of such companies.
The market is small, but it is lucrative as researchers tend to work with larger clusters individually.

Also, we target organizations that are low on devops resources.
A biomedical lab is willing to spend extra money on more experienced scientists.
The final goal is the following.
If your professionals can use tools like Microsoft Access, or Excel, you will be able to use this one as well.

Please consult with a professional of your local jurisdiction.

## Getting started

```
git clone https://gitlab.com/eper.io/engine.git
```

## License

```
This document is Licensed under Creative Commons CC0.
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
to this document to the public domain worldwide.
This document is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this document.
If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.
```

