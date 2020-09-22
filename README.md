# cnetstat: a container-aware netstat
`cnetstat` dumps a list of TCP connections on a host, with their
Kubernetes container and pod names if they are from a container. It
currently assumes that the containers run on Docker, with labels in
the format that my version of Kubelet uses.

Use it like this (from the repository root directory):
```
go build
sudo ./cnetstat
```

You should see output like this:
```
Namespace  Pod       Container    Protocol  Local Host        Local Port  Remote Host  Remote Port  Connection State
myapp      frontend  fe-server    https     aks-nodepool1-23  4592        10.2.9.76    https        ESTABLISHED
myapp      backend   be-server    https     aks-nodepool1-23  6820        10.2.10.82   https        ESTABLISHED
myapp      backend   log-scraper  https     aks-nodepool1-23  7819        10.2.9.83    https        TIME_WAIT
```

If you want JSON output, try this:
```
sudo ./cnetstat --format=json
```

# Why cnetstat?
We built cnetstat to help figure out which containers in a Kubernetes
cluster were using up TCP ports by opening lots of short-lived
outbound connections.

You might want to use cnetstat if you have that problem, or any
problem related to the interaction of container-level and host-level
networking. We hope cnetstat will be helpful to you too.

# Design and Roadmap

See the [design doc](https://github.com/microsoft/cnetstat/Design.md)

# Getting Involved
Is there a feature that would make cnetstat more useful for you? Are
you hitting a bug? Is the documentation unclear or lacking? Please let
us know!

See the [contributing
doc](https://github.com/microsoft/cnetstat/Contributing.md) for the
details.

Whether it's a feature, a bug report, or anything else, your
contributions make cnetstat better for everyone. Thank you.
