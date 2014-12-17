Deploying MailHog
=================

### Command line

You can run MailHog locally from the command line.

    go get github.com/ian-kent/MailHog/MailHog
    MailHog

To configure MailHog, use the environment variables or command line flags
described in the [README](README.md).

### Using supervisord/upstart/etc

MailHog can be started as a daemon using supervisord/upstart/etc.

See [this example init script](https://github.com/geerlingguy/ansible-role-mailhog/blob/master/files/mailhog)
and [this Ansible role](https://github.com/geerlingguy/ansible-role-mailhog) by [geerlingguy](https://github.com/geerlingguy).

### Docker

The example [Dockerfile](Dockerfile) can be used to run MailHog in a [Docker](https://www.docker.com/) container.

### Elastic Beanstalk

You can deploy Go-MailHog using [AWS Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/).

1. Open the Elastic Beanstalk console
2. Create a zip file containing the Dockerfile and MailHog binary
3. Create a new Elastic Beanstalk application
4. Launch a new environment and upload the zip file

If you're using in-memory storage, you can only use a single instance of
Go-MailHog. To use a load balanced EB application, use MongoDB backed storage.

To configure your Elastic Beanstalk MailHog instance, either:

* Set environment variables using the Elastic Beanstalk console
* Edit the Dockerfile to pass in command line arguments
