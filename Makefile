## ----------------------------------------------------------------------
## This makefile can be used to execute common functions to interact with
## the source code, these functions ease local development and can also be
## used in CI/CD pipelines.
## ----------------------------------------------------------------------

docker_args=
docker_scale_num=2

# REFERENCE: https://stackoverflow.com/questions/16931770/makefile4-missing-separator-stop
help: ## - Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

build: ## - build the source (latest)
	@docker ${docker_args} compose --profile client build --build-arg GIT_COMMIT=`git rev-parse HEAD` \
	--build-arg GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
	@docker ${docker_args} image prune -f

run: ## - run the service and its dependencies (docker) detached
	@docker ${docker_args} compose up -d --wait --scale server=${docker_scale_num}

stop:
	@docker ${docker_args} compose --profile client down

clean:
