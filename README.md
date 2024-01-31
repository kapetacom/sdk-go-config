# Config SDK for Go

## Description

This is a small library for getting configuration for application running in-side of Kapeta.

There are two providers supported.

### Local provider

This provider is used when running Kapeta locally i.e. via the desktop / command line or IDE.
This talks with the [Local Cluster Service](https://github.com/kapetacom/local-cluster-service/) to get informations
about blocks and plans.

### Kubernetes provider

This is used when running the block in Kubernetes, the provider is configured via environment variables.
These are injected in the the container when using the Kapeta deployment targets.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details