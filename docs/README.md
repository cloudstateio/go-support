# Cloudstate Go documentation

Documentation source for Cloudstate Go, published to https://cloudstate.io/docs/go/current/

To build the docs with [sbt](https://www.scala-sbt.org):

```
sbt paradox
```

Can also first start the sbt interactive shell with `sbt`, then run commands.

The documentation can be viewed locally by opening the generated pages:

```
open target/paradox/site/main/index.html
```

To watch files for changes and rebuild docs automatically:

```
sbt ~paradox
```
