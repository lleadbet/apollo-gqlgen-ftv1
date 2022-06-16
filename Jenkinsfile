pipeline {
  agent {
    docker {
      image 'golang:1.17.5-alpine'
    }

  }
  stages {
    stage('Rover Install') {
      steps {
        retry(count: 2) {
          sh '''apk add --update \\
    curl \\
    && rm -rf /var/cache/apk/*
curl -sSL https://rover.apollo.dev/nix/latest | sh'''
        }

      }
    }

    stage('Rover Checks') {
      steps {
        sh 'rover --help'
      }
    }

  }
}