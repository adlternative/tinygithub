.PHONY: build clean

build:
	mkdir -p build && \
	cd build && \
	git clone --depth=1 https://github.com/adlternative/tinygithub-frontend.git && \
	cd tinygithub-frontend && \
	npm install && \
	VUE_APP_GIT_CLONE_URL=localhost VUE_APP_BASE_HOST=localhost npm run build

clean:
	rm -rf build