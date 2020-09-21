# cnetstat design

The cnetstat data processing pipeline looks like this:
1. Use `lsns` to get a list of all the net namespaces we can see.
1. Use `nsenter -t <pid> -n netstat` to get a list of connections in
   each namespace, with their PID.
1. Use `docker` to get a map from PIDs to Docker container labels,
   which include the Kubernetes namespace, pod, and container name.
1. Match the PIDs from netstat with the PIDs from Docker, yielding a
   list of connections with their container identifiers.

## Pid-to-pod transation
One of the key things cnetstat needs to do is translate host PIDs to
Kubernetes container and pod names. I know of two ways of doing this
pid-to-pod translation.

The Docker way:

1. Use `docker ps` to get the docker container ID and labels of every pod.
2. Iterate through all pods, using `docker inspect` to get the root PID of
   each one.

The cgroup way:

1. Iterate through /sys/fs/cgroup/cpu,cpuacct/kubepods/... to get all
   Kubernetes pods and containers on the system, with all of their PIDs and
   their Kubernetes UIDs.
2. Use `docker ps` to translate Kubernetes UIDs into Kubernetes container and
   pod names.

They both require iterating over all pods in the system. The cgroup way has
the advantage that it gives all PIDs in a pod, not just the root PID, and
also that we could cache known UIDs. The Docker way has the advantage that it
uses public interfaces, instead of implementation details.

I'm using the Docker way because it seems easier to get a proof-of-concept
with, but I'm not sure what will be better in the long term.

## Net namespaces
One important design point is that cnetstat builds its pid-to-pod
mapping by talking to Docker, but it doesn't just iterate through
connections from Docker-owned PIDs. Instead, it uses `lsns` to get a
list of all net namespaces on a host, gets all connections from all of
those namespaces, and then reports container identities for the PIDs
that have them.

This is how cnetstat lists all connections, including those from the
host or non-Docker container systems.

## Future goals

### Add an option to run in a loop and print connections periodically
This is required for some of the features below.

### Add an option to print summary statistics instead of a connection list

### Track the owning process of TIME_WAIT connections
When a process closes a connection, it goes into state
`TIME_WAIT`. netstat doesn't print an associated PID any more (likely
because the kernel doesn't consider it associated with a PID), but I
still want to know which process opened it so I can debug processes
that open lots of short-lived connections. We just need to keep our
own map from connections to PIDs and update it in the polling loop.

### Use socket open/close events directly
Instead of using netstat to get the list of open connections every
time, we should get the list of connections once, at startup, and then
get a stream of socket open/close events from the kernel. I didn't do
this initially because I wanted to get cnetstat working quickly, but I
think this is the right thing to do, both because the algorithmic
complexity will be better and because it will ensure that we can
attribute every connection to a process, even if it isn't open when we
poll open connections.

### Include a Kubernetes pod specification for running cnetstat as a daemonset

### Support non-Docker container runtimes
I don't have an example around to test with, but I would gladly accept
a pull request for pid-to-pod translation for a non-Docker container
runtime.
