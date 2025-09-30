migrate:
	go run ./cmd/terndotenv/ migrate

gen:
	go generate ./...

run:
	go run ./cmd/api/main.go

docs:
	swag init --generalInfo internal/api/api.go --dir ./ --parseInternal

makebin: gen
	@mkdir -p ./bin

buildapp: makebin
	GOOS=linux CGO_ENABLED=0 go build -o ./bin ./cmd/api/


GO=$(shell which go)
APP_NAME=gobid
REGION=${AWS_REGION}
VPC_ID=${AWS_VPC_ID}
SG_NAME=${APP_NAME}-sg
ACCOUNT_ID=${AWS_ACCOUNT_ID}
ECR_URL=${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com
REPO_URL=${ECR_URL}/${APP_NAME}
DB_NAME=gobid
ACCESS_ROLE_ARN=${AWS_ACCESS_ROLE_ARN}
DB_PASSWORD=${GOBID_DATABASE_PASSWORD}

create-sg:
	@if ! aws ec2 describe-security-groups --filters "Name=group-name,Values=${SG_NAME}" --region ${REGION} --query "SecurityGroups[*].GroupId" --output text | grep -qE 'sg-'; then \
		echo "Creating security group ${SG_NAME}..."; \
		aws ec2 create-security-group \
			--group-name ${SG_NAME} \
			--description "Allow Postgres for gobid" \
			--vpc-id ${VPC_ID} \
			--region ${REGION}; \
	else \
		echo "Security group already exists."; \
	fi; \
	SG_ID=$$(aws ec2 describe-security-groups --filters "Name=group-name,Values=${SG_NAME}" --region ${REGION} --query "SecurityGroups[*].GroupId" --output text); \
	echo "Authorizing port 5432 on SG $$SG_ID..."; \
	aws ec2 authorize-security-group-ingress \
		--group-id $$SG_ID \
		--protocol tcp \
		--port 5432 \
		--cidr 0.0.0.0/0 \
		--region ${REGION} || echo "Ingress rule already exists or failed."


create-ecr:
	aws ecr describe-repositories --region ${REGION} --repository-name ${APP_NAME} || \
	aws ecr create-repository --repository-name ${APP_NAME} --region ${REGION}


build:
	docker build -t ${APP_NAME} -f Dockerfile.prod .

tag: build
	docker tag ${APP_NAME}:latest ${REPO_URL}:latest

push: tag
	@mkdir -p /tmp/docker-config
	DOCKER_CONFIG=/tmp/docker-config aws ecr get-login-password --region ${REGION} | \
	DOCKER_CONFIG=/tmp/docker-config docker login --username AWS --password-stdin ${ECR_URL}
	DOCKER_CONFIG=/tmp/docker-config docker push ${REPO_URL}:latest

create-db:
	@if ! aws rds describe-db-instances --db-instance-identifier $(DB_NAME) --region $(REGION) > /dev/null 2>&1; then \
		echo "Creating RDS Instance $(DB_NAME)..."; \
		SG_ID=$$(aws ec2 describe-security-groups --filters "Name=group-name,Values=${SG_NAME}" --region ${REGION} --query "SecurityGroups[*].GroupId" --output text); \
		aws rds create-db-instance \
			--db-instance-identifier $(DB_NAME) \
			--db-instance-class db.t3.micro \
			--engine postgres \
			--allocated-storage 20 \
			--master-username postgres \
			--master-user-password ${DB_PASSWORD} \
			--vpc-security-group-ids $$SG_ID \
			--publicly-accessible \
			--backup-retention-period 0 \
			--no-multi-az \
			--engine-version 17 \
			--port 5432 \
			--region $(REGION); \
	else \
	  echo "DRS instace $(DB_NAME) already exists.";\
	fi

create-schema:
	@DB_ENDPOINT=$$(aws rds describe-db-instances --db-instance-identifier $(DB_NAME) --region $(REGION) --query "DBInstances[0].Endpoint.Address" --output text); \
	PGPASSWORD=${DB_PASSWORD} psql -h $$DB_ENDPOINT -p 5432 -U postgres -c "CREATE DATABASE gobid;"

pingdb:
	 psql -h 0.0.0.0 -p 5432 -U gobid -d gobid -c "\l"

migration:
	$(GO) run ./cmd/terndotenv/ migrate

deploy:
	@if aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)']" --output text | grep -q '$(APP_NAME)'; then \
		echo "Service '$(APP_NAME) already exists. Updating ..."; \
		SERVICE_ARN=$$(aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)'].ServiceArn" --output text); \
		aws apprunner update-service \
			--service-arn $$SERVICE_ARN \
			--source-configuration file://apprunner-config.json \
			--region $(REGION); \
	else \
	  echo "Service '$(APP_NAME)' does not exists. Creating ..."; \
	  aws apprunner create-service \
	  		--service-name $(APP_NAME) \
			--source-configuration  file://apprunner-config.json \
			--region $(REGION); \
	fi

shutdown:
	@echo "🛑 Shutting down all AWS services to save costs..."
	@echo "Pausing AppRunner service..."
	@SERVICE_ARN=$$(aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)'].ServiceArn" --output text 2>/dev/null); \
	if [ ! -z "$$SERVICE_ARN" ]; then \
		aws apprunner pause-service --service-arn $$SERVICE_ARN --region $(REGION) && \
		echo "✅ AppRunner service paused"; \
	else \
		echo "ℹ️  No AppRunner service found"; \
	fi
	@echo "Stopping RDS instance..."
	@aws rds stop-db-instance --db-instance-identifier $(DB_NAME) --region $(REGION) 2>/dev/null && \
	echo "✅ RDS instance stopped" || echo "ℹ️  RDS instance not found or already stopped"
	@echo "🎯 All services shutdown complete!"

startup:
	@echo "🚀 Starting up AWS services..."
	@echo "Starting RDS instance..."
	@aws rds start-db-instance --db-instance-identifier $(DB_NAME) --region $(REGION) 2>/dev/null && \
	echo "✅ RDS instance starting" || echo "ℹ️  RDS instance not found or already running"
	@echo "Resuming AppRunner service..."
	@SERVICE_ARN=$$(aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)'].ServiceArn" --output text 2>/dev/null); \
	if [ ! -z "$$SERVICE_ARN" ]; then \
		aws apprunner resume-service --service-arn $$SERVICE_ARN --region $(REGION) && \
		echo "✅ AppRunner service resumed"; \
	else \
		echo "ℹ️  No AppRunner service found"; \
	fi
	@echo "🎯 All services startup complete!"

status:
	@echo "📊 Checking AWS services status..."
	@echo "AppRunner status:"
	@aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)'].[ServiceName,Status]" --output table 2>/dev/null || echo "  No AppRunner services found"
	@echo ""
	@echo "RDS status:"
	@aws rds describe-db-instances --db-instance-identifier $(DB_NAME) --region $(REGION) --query "DBInstances[0].[DBInstanceIdentifier,DBInstanceStatus]" --output table 2>/dev/null || echo "  No RDS instances found"
	@echo ""
	@echo "ECR repositories:"
	@aws ecr describe-repositories --region $(REGION) --query "repositories[?repositoryName=='$(APP_NAME)'].[repositoryName,createdAt]" --output table 2>/dev/null || echo "  No ECR repositories found"

clean-resources:
	@echo "⚠️  WARNING: This will DELETE all AWS resources!"
	@echo "This action cannot be undone. Press Ctrl+C to cancel."
	@read -p "Type 'DELETE' to confirm: " confirm; \
	if [ "$$confirm" = "DELETE" ]; then \
		echo "🗑️  Deleting AppRunner service..."; \
		SERVICE_ARN=$$(aws apprunner list-services --region $(REGION) --query "ServiceSummaryList[?ServiceName=='$(APP_NAME)'].ServiceArn" --output text 2>/dev/null); \
		if [ ! -z "$$SERVICE_ARN" ]; then \
			aws apprunner delete-service --service-arn $$SERVICE_ARN --region $(REGION); \
		fi; \
		echo "🗑️  Deleting RDS instance..."; \
		aws rds delete-db-instance --db-instance-identifier $(DB_NAME) --skip-final-snapshot --region $(REGION) 2>/dev/null; \
		echo "🗑️  Deleting ECR repository..."; \
		aws ecr delete-repository --repository-name $(APP_NAME) --force --region $(REGION) 2>/dev/null; \
		echo "✅ All resources deleted!"; \
	else \
		echo "❌ Deletion cancelled"; \
	fi