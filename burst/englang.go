package burst

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is a module code that runs burst containers.
// Containers typically run as a docker container or other isolated process.
// They can be implemented as a co-located container in the same pod
// They communicate through local 127.0.0.1:2121 UDP calls
// This makes them safer
// Security design considerations.
//
// The locality is ensured by private keys distributed early starting with the activation key.
// This ensures that we have a local runner and user code cannot get the bust call of other user code.
// What does this mean?
// - Idle process is exposed on 127.0.0.1:2121, and it responds to local endpoints only.
// - Idle process returns a task and a key to complete the task.
// - Malicious tasks may go for idle again.
// - We protect against this by letting bursts run for a term e.g. ten seconds
// - We protect against this also by not issuing a new key until the previous one finishes. //TODO
// - Each runner connects to the Idle process using the activation key
// - It returns a unique burst session key for the container.
// - The activation key is deleted from the container once used. //TODO
// - The init task of the container is our burst runner. It should not be set debuggable by workload.
// - Do not use secrets, these can be read by the Idle process stub, but by workload as well.
// - The runner restarts after each run, so that any local state and code is lost disabling double /idle calls.
// - The init task also terminates the container, if the workload tries to kill it.
// - The final column is time fencing allowing /idle calls only once every minute when workloads are already gone. //TODO
