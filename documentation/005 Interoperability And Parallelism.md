# Interoperability And Parallelism
### Author: Creative Commons Zero
### Date: 2023-05-29

## Critique

You may wonder with so many random isolated containers, whether this will ever work in large.
Here are the design considerations that will make it work.

**Problem** High performance clients may retrieve data at a speed that entries propagate slower.
Other streaming databases like Flink, Kafka, Spark have similar issues.

**Solution** The typical solution is to set an eventually consistent deadline.
The reason this is ten seconds, so that two team members should be able to verify the integrity.
Refreshing a browser window within ten seconds is not such a big deal.

**Problem** Indexed servers may be out of sync causing delays on office clusters larger than hundred nodes.

**Solution** The correct algorithm wakes up every ten seconds or so and sends the index to the neighbour.
A maximum cluster delay can be set by restricting the index width to eight characters for example.
An even better solution is to wake up and forward the index right away,
if we get a new refresh after at least six seconds or so have elapsed since the last forward refresh.

**Problem** Containerization is always hard.
How do you solve cases when a burst tries to steal a new burst lying to be idle?

**Solution**
The solution is simple. Any container should go to the local host service reporting it is idle.
They can get a newly generated key, and the burst server should record the time stamp.
If we return code to an idle container we choose with a key that has been generated more than ten seconds ago.
If the burst is restricted to run for three seconds, no burst key can be used so early.
All we need to make sure is to have enough idle containers around.

This is a clean time fencing solution.
Keys are regenerated once containers restarted, making any interim keys useless.



