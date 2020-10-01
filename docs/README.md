# Cloudstate Go documentation

The Cloudstate documentation is built using [Antora](https://antora.org) with Asciidoc sources.

The build is defined in the [Makefile](Makefile) and requires `make`, `bash`, and `docker`.

To build the documentation run:

```
make
```

The generated documentation site will be available in the `build/site` directory:

```
open build/site/index.html
```

Documentation will be automatically deployed on tagged versions, in the Travis CI builds.
