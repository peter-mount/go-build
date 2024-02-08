properties([
  buildDiscarder(
    logRotator(
      artifactDaysToKeepStr: '',
      artifactNumToKeepStr: '',
      daysToKeepStr: '',
      numToKeepStr: '10'
    )
  ),
  disableConcurrentBuilds(),
  disableResume(),
  pipelineTriggers([
    cron("H H * * *")
  ])
])
node("go") {
  stage("Checkout") {
    checkout scm
  }
  stage("Init") {
    sh 'make clean init'
  }
  stage("Test") {
    sh 'make test'
  }
  stage("Build") {
    sh 'make -f Makefile.gen all'
  }
}