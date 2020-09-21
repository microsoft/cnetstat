# cnetstat: a container-aware netstat
`cnetstat` dumps a list of TCP connections on a host, with their
Kubernetes container and pod names if they are from a container. It
currently assumes that the containers run on Docker, with labels in
the format that my version of Kubelet uses.

Use it like this:
```
sudo cnetstat
```

You should see a list of processes, with their TCP state and the
container and pod that started them.

If you want JSON output, try this:
```
sudo cnetstat --format=json
```

# Why cnetstat?
I built cnetstat to help me figure out which containers in a
Kubernetes cluster were using up TCP ports by opening lots of
short-lived outbound connections.

If you also have that problem, or any problem relating to the
interaction of container-level and host-level networking, then I hope
cnetstat will be helpful to you too.

# Design and Roadmap

See the [design doc](https://github.com/microsoft/cnetstat/design.md)

# Getting Involved
Is there a feature that would make cnetstat more useful for you? Are
you hitting a bug? Is the documentation unclear or lacking? Please let
me know!

For now, if you need to communicate about anything, please open an
issue on GitHub. Code contributions are very welcome, and must follow
the Microsoft CLA process in the Contributing section below. However,
don't feel like you have to write code to contribute - all ideas and
bug reports are welcome.

Whether it's a feature, a bug report, or anything else, your
contributions make cnetstat better for everyone. Thank you.

# Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
