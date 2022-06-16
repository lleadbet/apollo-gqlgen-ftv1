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
          sh '''apt-get update && apt-get install. -y curl
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