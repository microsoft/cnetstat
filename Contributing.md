# Contributing to cnetstat
Would you like to contribute to cnetstat? We'd love to have you on
board.

## Communication
For now, if you need to communicate about anything, please open an
issue on GitHub.

Code contributions are very welcome, and must follow the Microsoft CLA
process in the Contributing section below. However, don't feel like
you have to write code to contribute - all ideas and bug reports are
welcome.

## Contribution process
The process for changing the contents of this repository is:
1. Fork the repository to your own GitHub account
1. Make changes in your own copy of the repository
1. Send the changes as a GitHub pull request

Every pull request should be reviewed and approved by at least one
engineer in the cnetstat team before being merged.

Before you make a change, please, file an issue and talk with us about
what you want to do.

## Engineering
To build cnetstat, run
```
go build
```

in the project root directory. You can run tests like this:
```
go test
```

cnetstat depends on having `lsns`, `nsenter`, and `netstat` available.

## Code of Conduct
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Contributor License Agreement
This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.
