.PHONY: live fmt tidy gen css build run clean live/templ live/server live/tailwind live/sync_assets
TMP_DIR := tmp
STATIC_DIR := static
AIR_VERSION := 1.63.0

fmt:
	@gofmt -w .

tidy:
	@go mod tidy

gen:
	@templ generate ./...

css:
	@mkdir -p $(STATIC_DIR)
	@tailwindcss -i ./input.css -o ./$(STATIC_DIR)/output.css --minify

# UI DEVELOPMENT
# run templ generation in watch mode to detect all .templ files and
# recreate files on change, then send reload event to browser
# start all 4 watch processes in parallel
live:
	@mkdir -p $(TMP_DIR) $(STATIC_DIR)
	@echo "Starting full hot-reload environment"
	@$(MAKE) -j4 live/templ live/tailwind live/server live/sync_assets

live/templ:
	@echo "[LIVE] Starting templ proxy on :7331 (proxying to :8080)"
	@templ generate --watch --proxy="http://localhost:8080" --open-browser=false

# run air to detect any go file changes to re-build and re-run the server
live/server:
	@echo "[LIVE] Starting air server watcher..."
	@go run github.com/air-verse/air@v$(AIR_VERSION) \
	--build.cmd "go build -o ./$(TMP_DIR)/app ./cmd/dashboard" \
	--build.bin "./$(TMP_DIR)/app" \
	--build.full_bin "APP_ENV=dev ./$(TMP_DIR)/app serve" \
	--build.delay "100" \
	--build.exclude_dir "$(STATIC_DIR),$(TMP_DIR),migrations" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate output.css bundle in watch mode
live/tailwind:
	@echo "[LIVE] Starting tailwindcss watcher..."
	@mkdir -p $(STATIC_DIR)
	@tailwindcss -i ./input.css -o ./$(STATIC_DIR)/output.css --minify --watch=always


# watch for any js or css changes in the static/ folder, then reload browswer via templ proxy
live/sync_assets:
	@echo "[LIVE] Starting static asset watcher..."
	@go run github.com/air-verse/air@v$(AIR_VERSION) \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "/bin/true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "$(STATIC_DIR)" \
	--build.include_ext "js,css"

# Generate templates/CSS and build binary
build: gen css
	@mkdir -p $(TMP_DIR)
	@go build -o ./$(TMP_DIR)/app ./cmd/dashboard

run: build
	@./$(TMP_DIR)/app serve

clean:
	@rm -rf ./$(TMP_DIR)