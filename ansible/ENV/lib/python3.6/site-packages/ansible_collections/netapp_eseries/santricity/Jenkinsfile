@Library("hub")_
setupBlackduckBuildParameters()

def hubProjectName = "ansible-santricity-collection"  //"esg"
def hubProjectVersion = "1.2"


pipeline {
    agent {
        label "linux-docker"
    }

    options {
        timestamps()
        timeout(time: 60, unit: "MINUTES")
    }

    stages {
        stage("Hub scan") {
            // NOTE: Using the declarative 'agent { docker image ...}' tends to run on a different node. So don't use it here.
            steps {
                script {
                    docker.image("docker.netapp.com/mswbuild/openjdk8:8u181-8").inside {
                        hubScanProject(
                            "${hubProjectName}",
                            "${hubProjectVersion}",
                            productionScan: true
                        )
                    }
                }
            }
        }
    }

    post {
        always {
            deleteDir()
        }
    }
}
