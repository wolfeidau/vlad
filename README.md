# vlad

A versatile lightweight application deployer (VLAD), this is modeled on the idea of using a declaritive YAML based format to automate the deployment of applications. It leans on the standard templating build into Go, along with the [Amazon Web Services Go SDK](https://aws.amazon.com/) to enable deployment of AWS resources.

# Overview

Things I am planning to build into this tool:

* Make something fast, simple and focused on a core set of problems.
* Make it easy to get started with, nice to use, and provide visibility into the execution.
* Wrangle vars for multiple environments.
* Create, configure and delete resources in the [AWS](http://aws.amazon.com/) cloud platform.
* Enable tight integration with cloud services for configuration storage and encryption.
* Integrate metrics and flight recording into the execution process to support testing.
* Make testing tasks a first class citizen using recorded data playback to perform end to end testing of a [RunBook](#runbook).

The things I am NOT planning to do in this project:

* Rebuild Ansible
* Make a massive everything tool
* SSH to anything

# Usage

Given a simple YAML [RunBook](#runbook) file from the examples folder.

```yaml
vars:
    env: dev
    env_no: 1

tasks:
    - name: launch a lambda cloudformation stack
      cloudformation:
        stack_name: "deploy-lambda-{.env}-{.env_no}"
        template: "cloudformation/lambda.yml"
        disable_rollback: true
        tags:
            Owner: "wolfeidau"
```

# Features

* Execute a RunBook with some embedded vars.
* Launch and wait for the completion of [cloudformation](https://aws.amazon.com/cloudformation/) stacks.

## RunBook

A RunBook is a declarative YAML file which describes a list of vars, and some tasks to execute in sequence.

# Become a Contributor

vlad is provided free of cost because of the contributions that are made from developers like you. If you'd like to see this project grow, we would love it if you could submit a pull request to the project on GitHub.

# License 

Copyright 2018 Mark Wolfe. This project is licensed under the Apache License 2.0.