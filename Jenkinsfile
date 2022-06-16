pipeline {
  agent {
    dockerfile {
      filename 'golang:1.17.5-alpine'
    }

  }
  stages {
    stage('Rover Install') {
      steps {
        retry(count: 2) {
          sh 'curl -sSL https://rover.apollo.dev/nix/latest | sh'
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