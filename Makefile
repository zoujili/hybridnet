REGISTRY=github/oecp
ARCHS?=amd64
DEV_TAG?=dev_xizhe
RELEASE_TAG?=release

.PHONY: build-dev-images release

build-dev-images:
	@for arch in ${ARCHS} ; do \
    	docker build -t ${REGISTRY}/rama:${DEV_TAG}-$$arch -f Dockerfile.$$arch ./; \
    done

release:
	@for arch in ${ARCHS} ; do \
		docker build -t ${REGISTRY}/rama:${RELEASE_TAG}-$$arch -f Dockerfile.$$arch ./; \
	done
