#!/usr/bin/env bash
#
# DESCRIPTION: ECS Deployment Script
# MAINTAINER: Mariano Casero
# DEPENDENCIES: bash (>= 4.4.12), python (~> 2.7.13), awscli (~> 1.11.91), docker-ce (>= 17.03.1)
# FROM: https://github.com/atomicobject/ecs-deployment
#

export PATH=~/.local/bin:$PATH

set -e

# BEGIN CUSTOMIZATIONS #
ECS_REGION='us-east-1'
ECS_CLUSTER_NAME='Eco'
ECS_SERVICE_NAME='SRV-Eco-Search-PRODUCTION'
ECS_TASK_DEFINITION_NAME='Eco-Search-PRODUCTION'
ECR_NAME='eco/back'
ECR_URI='798873886583.dkr.ecr.us-east-1.amazonaws.com'
VERSION=$(date +%s)
AWSCLI_VER_TAR=1.11.91
# END CUSTOMIZATIONS #

# BEGIN OTHER VAR DEFINITIONS #
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ORIGINAL_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
ENVIRONMENT=""
BRANCH=""
AWSCLI_VER=$(aws --version 2>&1 | cut -d ' ' -f 1 | cut -d '/' -f 2)
# END OTHER VAR DEFINITIONS #

if [[ ${AWSCLI_VER} < ${AWSCLI_VER_TAR} ]]
then echo "ERROR: Please upgrade your AWS CLI to version ${AWSCLI_VER_TAR} or later!"
  exit 1
fi

ENVIRONMENT="PRODUCTION"
BRANCH="master"


echo "You are deploying ${BRANCH} to ${ENVIRONMENT}."

# Git operations
(

  # Fetch latest from git origin
  git fetch --all

  # Checkout the git branch
  git checkout "${BRANCH}"
  
  # actualizo dependencias govendor
  # solo si tengo configurada la carpeta vendor
  #if [ -d "vendor" ]; then
  #  govendor update +vendor
  #fi
)

#Compilo la ultima version
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o application . 

#Como es test, uso el yaml de test
mv server.yml server.yml.bak
mv server.yml.PRODUCTION server.yml


DOCKERIMAGENAME="search-production"
# Docker operations
(
    
  # Build the Docker image (to do asset and template compilation, etc.)
  sudo docker build -t "${DOCKERIMAGENAME}:latest" -f dockerfile .

  # Tag the new Docker image to the remote repo (by date)
  sudo docker tag "${DOCKERIMAGENAME}:latest" "${ECR_URI}/${ECR_NAME}:${DOCKERIMAGENAME}-${VERSION}"

  # Login to ECR us-east-1
  sudo $(aws ecr get-login --region "${ECS_REGION}" --no-include-email)

  # Push to the remote repo (by date)
  sudo docker push "${ECR_URI}/${ECR_NAME}:${DOCKERIMAGENAME}-${VERSION}"

)

#vuelvo a dejar el server.yaml de test
mv server.yml server.yml.PRODUCTION
mv server.yml.bak server.yml


# ECS operations
(
  
  # Store revision
  REVISION=$(git rev-parse "${BRANCH}")

  # Get previous task definition from ECS
  PREVIOUS_TASK_DEF=$(aws ecs describe-task-definition --region "${ECS_REGION}" --task-definition "${ECS_TASK_DEFINITION_NAME}")

  # Create the new ECS container definition from the last task definition
  NEW_CONTAINER_DEF=$(echo "${PREVIOUS_TASK_DEF}" | python <(cat <<-EOF
import sys, json
definition = json.load(sys.stdin)['taskDefinition']['containerDefinitions']
definition[0]['image'] = '${ECR_URI}/${ECR_NAME}:${DOCKERIMAGENAME}-${VERSION}'
print json.dumps(definition)
EOF
  ))

  # Create the new task definition
  aws ecs register-task-definition --region "${ECS_REGION}" --family "${ECS_TASK_DEFINITION_NAME}" --container-definitions "${NEW_CONTAINER_DEF}" --task-role-arn "arn:aws:iam::798873886583:role/ecsTaskExecutionRole"

  # Update the service to use the new task defintion
  aws ecs update-service --region "${ECS_REGION}" --cluster "${ECS_CLUSTER_NAME}" --service "${ECS_SERVICE_NAME}" --task-definition "${ECS_TASK_DEFINITION_NAME}"
  
  # Stop previous task
  #TODO stopear la tarea anterior
  #PREVIOUS_TASK_ARN=$(aws ecs describe-task-definition --region "us-east-1" --task-definition "EcoBackEmail" | awk -F ' ' '{print $1 " " $24}' | cut -c6-66)
  #aws ecs stop-task --cluster "${ECS_CLUSTER_NAME}" --task "${PREVIOUS_TASK_ARN}"

)

# Return to original branch
(
  # Checkout the original branch
  git checkout "${ORIGINAL_BRANCH}"
)
