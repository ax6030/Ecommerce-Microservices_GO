pipeline {
    agent any

    parameters {
        string(name: 'IMAGE_TAG', defaultValue: 'develop', description: 'Docker image tag to deploy')
    }

    environment {
        REGISTRY     = 'ghcr.io'
        IMAGE_PREFIX = 'ghcr.io/ax6030/ecommerce'
        VM_HOST      = '192.168.132.128'
        VM_USER      = 'ubuntu'
        DEPLOY_DIR   = '/opt/ecommerce'
    }

    stages {
        stage('Checkout') {
            steps {
                git branch: 'develop',
                    url: 'https://github.com/ax6030/Ecommerce-Microservices_GO.git'
            }
        }

        stage('Copy compose file to VM') {
            steps {
                sshagent(['vm-ssh-credentials']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ${VM_USER}@${VM_HOST} \
                            "mkdir -p ${DEPLOY_DIR}"

                        scp -o StrictHostKeyChecking=no \
                            docker-compose.prod.yml \
                            ${VM_USER}@${VM_HOST}:${DEPLOY_DIR}/docker-compose.prod.yml
                    """
                }
            }
        }

        stage('Deploy to VM') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'ghcr-credentials',
                    usernameVariable: 'GHCR_USER',
                    passwordVariable: 'GHCR_TOKEN'
                )]) {
                    sshagent(['vm-ssh-credentials']) {
                        sh """
                            ssh -o StrictHostKeyChecking=no ${VM_USER}@${VM_HOST} bash << 'REMOTE'
                                set -e
                                cd ${DEPLOY_DIR}

                                # Login to GHCR
                                echo "${GHCR_TOKEN}" | docker login ghcr.io -u "${GHCR_USER}" --password-stdin

                                # Pull latest images
                                export IMAGE_TAG=${IMAGE_TAG}
                                export IMAGE_PREFIX=${IMAGE_PREFIX}

                                for svc in user-service product-service order-service notification-service api-gateway; do
                                    docker pull ${IMAGE_PREFIX}-\${svc}:${IMAGE_TAG} || \
                                    docker pull ${IMAGE_PREFIX}-\${svc}:develop
                                done

                                # Rolling update
                                IMAGE_TAG=${IMAGE_TAG} IMAGE_PREFIX=${IMAGE_PREFIX} \
                                    docker compose -f docker-compose.prod.yml up -d --no-build --remove-orphans
REMOTE
                        """
                    }
                }
            }
        }

        stage('Health Check') {
            steps {
                sshagent(['vm-ssh-credentials']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ${VM_USER}@${VM_HOST} bash << 'REMOTE'
                            set -e
                            echo "Waiting for services..."
                            sleep 20

                            check() {
                                local url=\$1
                                local name=\$2
                                if curl -sf --retry 5 --retry-delay 3 "\$url" > /dev/null; then
                                    echo "✓ \$name healthy"
                                else
                                    echo "✗ \$name FAILED"
                                    exit 1
                                fi
                            }

                            check http://localhost:8080/healthz "api-gateway"
                            check http://localhost:8081/healthz "user-service"
                            check http://localhost:8082/healthz "product-service"
                            check http://localhost:8083/healthz "order-service"
REMOTE
                    """
                }
            }
        }
    }

    post {
        failure {
            echo "Deploy failed on VM ${VM_HOST}"
            sshagent(['vm-ssh-credentials']) {
                sh """
                    ssh -o StrictHostKeyChecking=no ${VM_USER}@${VM_HOST} \
                        "cd ${DEPLOY_DIR} && docker compose -f docker-compose.prod.yml down || true"
                """
            }
        }
        success {
            echo "Deploy succeeded: ${IMAGE_TAG} → ${VM_HOST}"
        }
    }
}
