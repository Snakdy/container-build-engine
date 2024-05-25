# Container Build Engine

Foundational components used to build application containers without privileges.

## How it works

CBE is similar to tools like Jib and Ko in that it builds a virtual file system (which may include an application, or applications) and then appends that to a "base" container image.

CBE uses the concept of a pipeline.
A pipeline contains a set of statements which execute a piece of logic.

An example of a statement includes:

* Setting an environment variable
* Downloading a file
* Creating a symbolic link

Custom statements can be included to add custom functionality (e.g. installing packages, building an executable application).

## Usage

Usage documentation can be found in the [`docs`](docs) directory.

## Reference implementation

CBE includes a reference implementation that builds a container from a YAML configuration file.
While it's not exactly production-ready, there's nothing stopping you from using it.

Example configuration files can be found in the [`fixtures/v1`](fixtures/v1) directory.
