# Make Cloudstate Go documentation

module   := go
upstream := cloudstateio/go-support
branch   := docs/current
sources  := src build/src/managed

cloudstate_antora_download := https://github.com/cloudstateio/cloudstate-antora/raw/master/cloudstate-antora
cloudstate_antora := .cache/bin/cloudstate-antora
descriptor := build/site.yml
root_dir := $(shell git rev-parse --show-toplevel)
src_managed := build/src/managed
managed_examples := ${src_managed}/modules/go/examples
managed_partials := ${src_managed}/modules/go/partials

define copy_example
	mkdir -p "${managed_examples}/$$(dirname $1)"
	cp "${root_dir}/$1" "${managed_examples}/$1"
endef

.SILENT:

build: clean managed validate html

${cloudstate_antora}:
	mkdir -p $$(dirname ${cloudstate_antora})
	curl -Lo ${cloudstate_antora} ${cloudstate_antora_download}
	chmod +x ${cloudstate_antora}

clean-cache:
	rm -rf .cache

update: clean-cache ${cloudstate_antora}

clean: ${cloudstate_antora}
	${cloudstate_antora} clean

managed: attributes examples
	mkdir -p "${src_managed}"
	cp src/antora.yml "${src_managed}/antora.yml"

attributes: ${cloudstate_antora}
	mkdir -p "${managed_partials}"
	${cloudstate_antora} version | xargs -0  printf ":cloudstate-go-lib-version: %s" \
		> "${managed_partials}/attributes.adoc"

examples:
	# also create empty go module to ignore example sources
	mkdir -p "${managed_examples}"
	touch "${managed_examples}/go.mod"
	$(call copy_example,cloudstate/event.go)
	$(call copy_example,cloudstate/eventsourced.go)
	$(call copy_example,protobuf/example/shoppingcart/persistence/domain.proto)
	$(call copy_example,tck/cmd/tck_shoppingcart/shoppingcart.go)

${descriptor}: ${cloudstate_antora}
	mkdir -p $$(dirname ${descriptor})
	${cloudstate_antora} source --preview --upstream ${upstream} ${sources} > build/source.yml
	${cloudstate_antora} site --preview --exclude ${module} build/source.yml > ${descriptor}

validate: ${descriptor}
	${cloudstate_antora} validate ${descriptor}

html: ${descriptor}
	${cloudstate_antora} build ${descriptor}

validate-links: ${cloudstate_antora}
	${cloudstate_antora} validate --no-xrefs --links

deploy: clean managed
	${cloudstate_antora} deploy --module ${module} --upstream ${upstream} --branch ${branch} ${sources}
