export CGO_ENABLED=0
all:
	ko build --platform=linux/arm64 --tags=$(TAG) --sbom=none ./cmd/outage_monitor/
