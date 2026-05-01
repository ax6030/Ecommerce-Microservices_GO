pipeline {
    agent any

    parameters {
        string(name: 'IMAGE_TAG', defaultValue: 'develop', description: 'Docker image tag to deploy')
    }

    environment {
        REGISTRY      = 'ghcr.io'
        IMAGE_PREFIX  = "ghcr.io/${env.GITHUB_REPO_OWNER}/ecommerce"
        COMPOSE_FILE  = 'docker-compose.prod.yml'
        SERVICES      = 'user-service product-service order-service notification-service api-gateway'
    }

    stages {
        stage('Checkout') {
            steps {
                git branch: 'develop',
                    url: 'https://github.com/ax6030/Ecommerce-Microservices_GO.git'
            }
        }

        stage('Pull Images') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'ghcr-credentials',
                    usernameVariable: 'GHCR_USER',
                    passwordVariable: 'GHCR_TOKEN'
                )]) {
                    sh '''
                        echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
                        for svc in $SERVICES; do
                            docker pull ${IMAGE_PREFIX}-${svc}:${IMAGE_TAG} || \
                            docker pull ${IMAGE_PREFIX}-${svc}:develop
                        done
                    '''
                }
            }
        }

        stage('Deploy') {
            steps {
                sh '''
                    export IMAGE_TAG=${IMAGE_TAG}
                    docker compose -f ${COMPOSE_FILE} up -d --no-build --remove-orphans
                '''
            }
        }

        stage('Health Check') {
            steps {
                sh '''
                    echo "Waiting for services to start..."
                    sleep 15

                    check() {
                        local url=$1
                        local name=$2
                        if curl -sf --retry 5 --retry-delay 3 "$url" > /dev/null; then
                            echo "✓ $name healthy"
                        else
                            echo "✗ $name FAILED"
                            exit 1
                        fi
                    }

                    check http://localhost:8080/healthz  "api-gateway"
                    check http://localhost:8081/healthz  "user-service"
                    check http://localhost:8082/healthz  "product-service"
                    check http://localhost:8083/healthz  "order-service"
                '''
            }
        }
    }

    post {
        failure {
            echo "Deploy failed — rolling back..."
            sh 'docker compose -f ${COMPOSE_FILE} down || true'
        }
        success {
            echo "Deploy succeeded: ${IMAGE_TAG}"
        }
    }
}
