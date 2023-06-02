.DEFAULT_GOAL := docker-compose
.PHONY: docker-build docker-clean frontend-build-done clean

docker-compose: | frontend-build
	docker-compose --env-file=./docker-compose.env build

frontend-build-done:
	@echo "Frontend build is already done"

frontend-build:
	@if [ -d "build/tinygithub-frontend" ]; then \
		make frontend-build-done; \
	else \
		mkdir -p build && \
		cd build && \
		git clone --depth=1 https://github.com/adlternative/tinygithub-frontend.git && \
		cd tinygithub-frontend && \
		npm install && \
		VUE_APP_GIT_CLONE_URL=localhost VUE_APP_BASE_HOST=localhost npm run build; \
	fi
clean:
	rm -rf build