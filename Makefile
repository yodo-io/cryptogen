REPO := $(shell git remote get-url origin | awk -F ':' '{ print $$2 }' | sed 's/\.git//')
REGISTRY ?= quay.io
TAG := latest-$(shell date +%s)

# build docker image
docker-build:
	docker build -t $(REPO):$(TAG) .

# push to docker
docker-push: docker-build
	docker tag $(REPO):$(TAG) $(REGISTRY)/$(REPO):$(TAG)
	docker push $(REGISTRY)/$(REPO):$(TAG)

# create your own charts/cryptogen/values-lab.yaml for this (not included)
deploy-lab: docker-push
	helm upgrade -i --set image.tag=$(TAG) -f charts/cryptogen/values-lab.yaml cryptogen charts/cryptogen
