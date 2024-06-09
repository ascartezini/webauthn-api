build:
	@echo "Building application"
	GOOS=linux GOARCH=amd64 go build -o bin/bootstrap
	@echo "Adding execute permissions to bin/bootstrap"
	chmod +x bin/bootstrap
	@echo "Zipping bin/bootstrap to bin/deployment.zip"
	cd bin && zip deployment.zip bootstrap
	@echo "Build complete"

deploy: build
	@echo "Deploying application"
	cd terraform && terraform apply -auto-approve
	@echo "Deploy complete"